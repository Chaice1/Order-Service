FROM golang:1.25.1-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app/

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o api_gateway ./cmd/api-gateway


FROM alpine:latest

WORKDIR /root/


COPY --from=builder ./app/api_gateway .
COPY --from=builder /app/migrations ./migrations

CMD ["./api_gateway"]