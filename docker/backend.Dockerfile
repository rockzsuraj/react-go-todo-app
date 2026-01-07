FROM golang:1.22-alpine

WORKDIR /app
COPY apps/api/go.mod apps/api/go.sum ./
RUN go mod download

COPY apps/api .
RUN go build -o server ./cmd/server

EXPOSE 8080
CMD ["./server"]