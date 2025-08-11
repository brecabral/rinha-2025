FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app ./cmd/server/main.go

FROM alpine:3.22

WORKDIR /app

COPY --from=builder /app/app .

ENTRYPOINT ["./app"]