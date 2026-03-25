# Stage 1: Сборка
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Копируем файлы зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем бинарный файл (CGO_ENABLED=0 для статической линковки)
RUN CGO_ENABLED=0 GOOS=linux go build -o /inventory-service ./cmd/server/main.go

# Stage 2: Финальный образ
FROM alpine:latest

WORKDIR /
# Копируем только исполняемый файл из сборщика
COPY --from=builder /inventory-service /inventory-service

# Открываем порт
EXPOSE 8080

# Запуск
CMD ["/inventory-service"]
