FROM golang:latest

WORKDIR /app
COPY . .

ENTRYPOINT ["./cmd/main.go"]