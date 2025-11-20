
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY core-service/go.mod core-service/go.sum ./

RUN go mod download

COPY core-service/ .

RUN CGO_ENABLED=0 GOOS=linux go build -o core-service ./cmd/main.go

# --- Stage 2: Runner ---
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/core-service .

EXPOSE 8080
CMD ["./core-service"]