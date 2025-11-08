# ---------------------------
# Stage 1: Build the Go binary
# ---------------------------
FROM golang:1.23.2-alpine AS builder

# Install dependencies
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum from root
COPY ../go.mod ../go.sum ./

# Download dependencies
RUN go mod download

# Copy the entire project source (from repo root)
COPY ../ ./

# Build the Go binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# ---------------------------
# Stage 2: Create runtime image
# ---------------------------
FROM alpine:latest

# Install minimal runtime dependencies
RUN apk --no-cache add ca-certificates curl

# Create non-root user
RUN adduser -D -s /bin/sh appuser

# Set working directory
WORKDIR /app

# Copy compiled binary and config
COPY --from=builder /app/main .
COPY --from=builder /app/config.json* ./

# Set permissions
RUN chown -R appuser:appuser /app

# Use non-root user
USER appuser

# Expose app port
EXPOSE 8081

# Healthcheck
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8081/api/v1/auth/health || exit 1

# Run app
CMD ["./main"]
