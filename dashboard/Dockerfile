# ── Stage 1: Build the Go binary ──
FROM golang:1.22-alpine AS builder
WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o gopanel .

# ── Stage 2: Minimal runtime ──
FROM alpine:3.19
RUN apk --no-cache add ca-certificates

COPY --from=builder /app/gopanel /usr/local/bin/gopanel
COPY --from=builder /app/static /static

EXPOSE 9000
ENTRYPOINT ["gopanel"]
