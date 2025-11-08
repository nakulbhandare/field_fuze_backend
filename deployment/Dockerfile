# Stage 1: Builder
FROM golang:1.23.2-alpine AS builder
RUN apk add --no-cache git ca-certificates
WORKDIR /app

# Copy mod files from context root
COPY go.mod go.sum ./
RUN go mod download

# Copy all project source
COPY . .

# Build the app
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Stage 2: Runtime
FROM alpine:latest
RUN apk --no-cache add ca-certificates curl
RUN adduser -D -s /bin/sh appuser
WORKDIR /app

# Copy binary and config
COPY --from=builder /app/main .
COPY --from=builder /app/config.json* ./
RUN chown -R appuser:appuser /app
USER appuser
EXPOSE 8081

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8081/api/v1/auth/health || exit 1

CMD ["./main"]
