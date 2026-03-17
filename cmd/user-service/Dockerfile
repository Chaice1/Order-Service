FROM golang:1.25.1-alpine AS builder


RUN apk add --no-cache git ca-certificates

WORKDIR /app/

COPY go.mod go.sum ./
RUN go mod download


COPY . .


RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o user_service ./cmd/user-service



FROM alpine:latest

WORKDIR /root/



COPY --from=builder ./app/user_service .
COPY --from=builder /app/migrations ./migrations

CMD ["./user_service"]