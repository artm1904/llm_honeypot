package HTTP

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/mariocandela/beelzebub/v3/parser"
	"github.com/mariocandela/beelzebub/v3/plugins"
	"github.com/mariocandela/beelzebub/v3/tracer"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

// httpStatusLineRegex matches a leading HTTP status line like "HTTP/1.1 200 OK".
var httpStatusLineRegex = regexp.MustCompile(`^HTTP/\d\.\d\s+(\d{3})(?:\s+[^\r\n]*)?\r?\n`)

// sanitizeLLMHTTPResponse strips a leaked HTTP status line and headers from an
// LLM-generated body. If present, the parsed status code and headers are merged
// into resp so they still take effect; otherwise resp is left as-is and only the
// body is updated.
func sanitizeLLMHTTPResponse(raw string, resp *httpResponse) {
	body := strings.TrimLeft(raw, " \t\r\n")

	statusMatch := httpStatusLineRegex.FindStringSubmatchIndex(body)
	if statusMatch != nil {
		if code, err := strconv.Atoi(body[statusMatch[2]:statusMatch[3]]); err == nil {
			resp.StatusCode = code
		}
		body = body[statusMatch[1]:]
	} else if !looksLikeHeaderLine(body) {
		resp.Body = raw
		return
	}

	// Consume header lines until a blank line or a non-header line.
	for {
		nl := strings.Index(body, "\n")
		if nl == -1 {
			break
		}
		line := strings.TrimRight(body[:nl], "\r")
		if line == "" {
			body = body[nl+1:]
			break
		}
		if !looksLikeHeaderLine(line) {
			break
		}
		if colon := strings.Index(line, ":"); colon > 0 {
			resp.Headers = append(resp.Headers, line)
		}
		body = body[nl+1:]
	}

	resp.Body = body
}

func looksLikeHeaderLine(s string) bool {
	colon := strings.Index(s, ":")
	if colon <= 0 {
		return false
	}
	for i := 0; i < colon; i++ {
		c := s[i]
		if !((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' || c == '_') {
			return false
		}
	}
	return true
}

type HTTPStrategy struct{}

type httpResponse struct {
	StatusCode int
	Headers    []string
	Body       string
}

func (httpStrategy HTTPStrategy) Init(servConf parser.BeelzebubServiceConfiguration, tr tracer.Tracer) error {
	serverMux := http.NewServeMux()

	serverMux.HandleFunc("/", func(responseWriter http.ResponseWriter, request *http.Request) {
		var matched bool
		var resp httpResponse
		var err error
		for _, command := range servConf.Commands {
			var err error
			matched = command.Regex.MatchString(request.RequestURI)
			if matched && command.Method != "" && !strings.EqualFold(request.Method, command.Method) {
				matched = false
			}
			if matched {
				resp, err = buildHTTPResponse(servConf, tr, command, request)
				if err != nil {
					log.Errorf("error building http response: %s: %v", request.RequestURI, err)
					resp.StatusCode = 500
					resp.Body = "500 Internal Server Error"
				}
				break
			}
		}
		// If none of the main commands matched, and we have a fallback command configured, process it here.
		// The regexp is ignored for fallback commands, as they are catch-all for any request.
		if !matched {
			command := servConf.FallbackCommand
			if command.Handler != "" || command.Plugin != "" {
				resp, err = buildHTTPResponse(servConf, tr, command, request)
				if err != nil {
					log.Errorf("error building http response: %s: %v", request.RequestURI, err)
					resp.StatusCode = 500
					resp.Body = "500 Internal Server Error"
				}
			}
		}
		setResponseHeaders(responseWriter, resp.Headers, resp.StatusCode)
		fmt.Fprint(responseWriter, resp.Body)

	})
	go func() {
		var err error
		// Launch a TLS supporting server if we are supplied a TLS Key and Certificate.
		// If relative paths are supplied, they are relative to the CWD of the binary.
		// The can be self-signed, only the client will validate this (or not).
		if servConf.TLSKeyPath != "" && servConf.TLSCertPath != "" {
			err = http.ListenAndServeTLS(servConf.Address, servConf.TLSCertPath, servConf.TLSKeyPath, serverMux)
		} else {
			err = http.ListenAndServe(servConf.Address, serverMux)
		}
		if err != nil {
			log.Errorf("error during init HTTP Protocol: %v", err)
			return
		}
	}()

	log.WithFields(log.Fields{
		"port":     servConf.Address,
		"commands": len(servConf.Commands),
	}).Infof("Init service: %s", servConf.Description)
	return nil
}

func buildHTTPResponse(servConf parser.BeelzebubServiceConfiguration, tr tracer.Tracer, command parser.Command, request *http.Request) (httpResponse, error) {
	resp := httpResponse{
		Body:       command.Handler,
		Headers:    command.Headers,
		StatusCode: command.StatusCode,
	}

	// Limit body read to 1MB to prevent DoS attacks
	bodyBytes, err := io.ReadAll(io.LimitReader(request.Body, 1024*1024))
	body := ""
	if err == nil {
		body = string(bodyBytes)
	}
	traceRequest(request, tr, command, servConf.Description, body)

	if request.RequestURI == "/login" && request.Method == "POST" {
		if strings.Contains(body, "username=admin") && strings.Contains(body, "password=digsi") && strings.Contains(body, "role=admin") {
			resp.StatusCode = 302
			resp.Headers = append(resp.Headers, "Location: /dashboard")
			resp.Headers = append(resp.Headers, "Set-Cookie: session=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyIjoiYWRtaW4ifQ.signature; Path=/; HttpOnly")
			resp.Body = "<html><head><meta http-equiv=\"refresh\" content=\"0;url=/dashboard\"></head><body>Login successful, redirecting...</body></html>"
			return resp, nil
		}
	}

	if command.Plugin == plugins.LLMPluginName {
		llmProvider, err := plugins.FromStringToLLMProvider(servConf.Plugin.LLMProvider)
		if err != nil {
			log.Errorf("error: %v", err)
			resp.Body = "404 Not Found!"
			return resp, err
		}

		llmHoneypot := plugins.BuildHoneypot(nil, tracer.HTTP, llmProvider, servConf)
		llmHoneypotInstance := plugins.InitLLMHoneypot(*llmHoneypot)
		command := fmt.Sprintf("Method: %s, RequestURI: %s, Body: %s", request.Method, request.RequestURI, body)

		completions, err := llmHoneypotInstance.ExecuteModel(command)
		if err != nil {
			resp.Body = "404 Not Found!"
			return resp, fmt.Errorf("ExecuteModel error: %s, %v", command, err)
		}
		sanitizeLLMHTTPResponse(completions, &resp)
	}
	return resp, nil
}

func traceRequest(request *http.Request, tr tracer.Tracer, command parser.Command, HoneypotDescription, body string) {
	host, port, _ := net.SplitHostPort(request.RemoteAddr)

	event := tracer.Event{
		Msg:             "HTTP New request",
		RequestURI:      request.RequestURI,
		Protocol:        tracer.HTTP.String(),
		HTTPMethod:      request.Method,
		Body:            body,
		HostHTTPRequest: request.Host,
		UserAgent:       request.UserAgent(),
		Cookies:         mapCookiesToString(request.Cookies()),
		Headers:         mapHeaderToString(request.Header),
		HeadersMap:      request.Header,
		Status:          tracer.Stateless.String(),
		RemoteAddr:      request.RemoteAddr,
		SourceIp:        host,
		SourcePort:      port,
		ID:              uuid.New().String(),
		Description:     HoneypotDescription,
		Handler:         command.Name,
	}
	// Capture the TLS details from the request, if provided.
	if request.TLS != nil {
		event.Msg = "HTTPS New Request"
		event.TLSServerName = request.TLS.ServerName
	}
	tr.TraceEvent(event)
}

func mapHeaderToString(headers http.Header) string {
	headersString := ""

	for key := range headers {
		for _, values := range headers[key] {
			headersString += fmt.Sprintf("[Key: %s, values: %s],", key, values)
		}
	}

	return headersString
}

func mapCookiesToString(cookies []*http.Cookie) string {
	cookiesString := ""

	for _, cookie := range cookies {
		cookiesString += cookie.String()
	}

	return cookiesString
}

func setResponseHeaders(responseWriter http.ResponseWriter, headers []string, statusCode int) {
	for _, headerStr := range headers {
		keyValue := strings.Split(headerStr, ":")
		if len(keyValue) > 1 {
			responseWriter.Header().Add(keyValue[0], keyValue[1])
		}
	}
	// http.StatusText(statusCode): empty string if the code is unknown.
	if len(http.StatusText(statusCode)) > 0 {
		responseWriter.WriteHeader(statusCode)
	}
}
