FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /trading-bot ./cmd/bot

CMD ["/trading-bot"]
