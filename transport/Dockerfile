# transport/Dockerfile
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Копируем go.mod и go.sum
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь код
COPY . .

# Собираем бинарник
RUN go build -o transport cmd/main.go

# Финальный образ
FROM alpine:latest

WORKDIR /root/

# Копируем бинарник из сборщика
COPY --from=builder /app/transport .

# Открываем порт
EXPOSE 8000

# Запускаем
CMD ["./transport"]