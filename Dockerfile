# Этап 1: Сборка (Compiling)
FROM golang:alpine AS builder

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux

WORKDIR /root/beelzebub

# КОПИРУЕМ твой локальный проект внутрь сборочного контейнера
# Вместо git clone мы берем файлы с диска
COPY . .

# Скачиваем зависимости и собираем
RUN go mod download
RUN go build -o main .

# ВАЖНО: Проверяем пути логов. T-Pot ожидает логи именно здесь.
# Если ты уже настроил это в своем конфиге, эту строку можно убрать.
# Но для надежности лучше оставить замену пути:
RUN sed -i "s#logsPath: ./log#logsPath: ./configurations/log/beelzebub.json#g" configurations/beelzebub.yaml

# Этап 2: Финальный образ (Минимальный размер)
FROM scratch

# 1. Сначала создаем рабочую директорию
WORKDIR /opt/beelzebub

# Копируем сертификаты (нужны для HTTPS/LLM)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Копируем собранный бинарник и конфиги
COPY --from=builder --chown=2000:2000 /root/beelzebub/main .
COPY --from=builder --chown=2000:2000 /root/beelzebub/configurations ./configurations

# Настройки запуска
WORKDIR /opt/beelzebub
USER 2000:2000
ENTRYPOINT ["./main"]
