FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY apps/api/go.mod apps/api/go.sum ./
RUN go mod download

COPY apps/api .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates wget
WORKDIR /root/

COPY --from=builder /app/server .

EXPOSE 8080
CMD ["./server"]