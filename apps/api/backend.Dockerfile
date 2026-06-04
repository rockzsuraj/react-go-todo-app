# Multi-stage build for Go backend
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/server

# Development stage with proper hot reload
FROM golang:1.24-alpine AS development

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Install air for hot reload (use older version compatible with Go 1.24)
RUN go install github.com/cosmtrek/air@v1.49.0

# Copy air configuration
COPY .air.toml .air.toml

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/ready || exit 1

# Run air for hot reload
CMD ["air"]

# Production stage
FROM alpine:latest AS production

RUN apk --no-cache add ca-certificates tzdata wget

WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/main .
COPY --from=builder /app/migrations ./migrations

# Make the binary executable
RUN chmod +x ./main

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/ready || exit 1

# Run the production binary
CMD ["./main"]
