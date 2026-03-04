# 1. Используем легковесный образ Alpine (весит ~300МБ против 800МБ+ у обычного)
FROM golang:1.25.1-alpine

# 2. Устанавливаем git и сертификаты (необходимы Alpine для загрузки библиотек по HTTPS)
# --no-cache не сохраняет индекс пакетов, экономя место
RUN apk add --no-cache git ca-certificates

WORKDIR /app

# 3. Кэшируем зависимости: сначала копируем только файлы модов
COPY go.mod go.sum ./
RUN go mod download

# 4. Копируем остальной исходный код
COPY . .

# 5. Сборка бинарника с важными флагами:
# CGO_ENABLED=0 — отключаем зависимости от системных библиотек C (делаем бинарник статическим)
# GOOS=linux — гарантируем сборку под Linux
# -ldflags="-s -w" — вырезаем отладочную информацию (уменьшаем вес файла на 20-30%)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o user_service ./cmd/user-service

# 6. Указываем порт (просто как документ для Docker)
EXPOSE 8080

# 7. Запускаем приложение
CMD ["./user_service"]