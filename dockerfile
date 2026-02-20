FROM golang:1.24.5 AS builder

WORKDIR /app

#install gcc for CGO
RUN apt-get update && apt-get install -y gcc

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=1
ENV GOOS=linux
ENV GOARCH=arm64

RUN go build -o server ./cmd/main.go

#final image
FROM debian:stable-slim

WORKDIR /app

COPY --from=builder /app/server .

EXPOSE 3000

CMD ["./server"]
