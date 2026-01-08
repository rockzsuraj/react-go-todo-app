FROM golang:1.25-alpine

RUN apk add --no-cache git

WORKDIR /app

# Copy ONLY backend module files
COPY apps/api/go.mod apps/api/go.sum ./
RUN go mod download

# Copy backend source
COPY apps/api .

# Build binary
RUN go build -o server ./cmd/server

EXPOSE 8080
CMD ["./server"]