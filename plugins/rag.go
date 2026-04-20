package plugins

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
)

type KnowledgeChunk struct {
	Title   string
	Content string
}

type RAGKnowledgeBase struct {
	Chunks []KnowledgeChunk
}

var (
	globalRAG *RAGKnowledgeBase
	ragOnce   sync.Once
)

// LoadKnowledgeBase reads .md files and splits them by '## ' headers
func LoadKnowledgeBase(dirPath string) {
	ragOnce.Do(func() {
		globalRAG = &RAGKnowledgeBase{}

		if dirPath == "" {
			return
		}

		err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && strings.HasSuffix(path, ".md") {
				contentBytes, err := os.ReadFile(path)
				if err != nil {
					log.Errorf("failed to read knowledge file %s: %v", path, err)
					return nil
				}
				content := string(contentBytes)
				// Basic splitting by "## " (markdown headers)
				parts := strings.Split(content, "\n## ")
				for i, part := range parts {
					text := part
					if i > 0 {
						text = "## " + part
					}
					// extract title
					lines := strings.SplitN(text, "\n", 2)
					title := ""
					if len(lines) > 0 {
						title = strings.TrimSpace(lines[0])
					}
					globalRAG.Chunks = append(globalRAG.Chunks, KnowledgeChunk{
						Title:   title,
						Content: strings.TrimSpace(text),
					})
				}
				log.Infof("Loaded knowledge from %s (%d chunks)", path, len(parts))
			}
			return nil
		})

		if err != nil {
			log.Errorf("failed to load knowledge base from %s: %v", dirPath, err)
		}
	})
}

// Search returns topK most relevant chunks based on keyword overlap
func Search(query string, topK int) []string {
	if globalRAG == nil || len(globalRAG.Chunks) == 0 {
		return nil
	}

	queryTokens := tokenize(query)
	if len(queryTokens) == 0 {
		return nil
	}

	type ScoreChunk struct {
		Chunk KnowledgeChunk
		Score int
	}

	scores := make([]ScoreChunk, 0, len(globalRAG.Chunks))

	for _, chunk := range globalRAG.Chunks {
		score := 0
		chunkText := strings.ToLower(chunk.Title + " " + chunk.Content)
		for _, token := range queryTokens {
			if strings.Contains(chunkText, token) {
				score++
			}
		}
		if score > 0 {
			scores = append(scores, ScoreChunk{Chunk: chunk, Score: score})
		}
	}

	// Sort high to low
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].Score > scores[j].Score
	})

	var results []string
	for i := 0; i < len(scores) && i < topK; i++ {
		results = append(results, scores[i].Chunk.Title+"\n"+scores[i].Chunk.Content)
	}

	return results
}

func tokenize(s string) []string {
	s = strings.ToLower(s)
	// replace some common delimiters
	s = strings.ReplaceAll(s, "_", " ")
	s = strings.ReplaceAll(s, "-", " ")
	s = strings.ReplaceAll(s, "/", " ")
	s = strings.ReplaceAll(s, ".", " ")
	s = strings.ReplaceAll(s, ",", " ")
	words := strings.Fields(s)
	
	// filter short words
	var filtered []string
	for _, w := range words {
		if len(w) > 3 {
			filtered = append(filtered, w)
		}
	}
	return filtered
}
