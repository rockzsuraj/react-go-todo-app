FROM golang:1.25-alpine

RUN apk add --no-cache git curl

WORKDIR /app

RUN go install github.com/air-verse/air@latest

COPY . .

WORKDIR /app/apps/api
RUN go mod download

WORKDIR /app

EXPOSE 8080
CMD ["air"]