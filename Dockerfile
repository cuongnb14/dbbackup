FROM golang:1.22.5 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app main.go

FROM alpine:latest

RUN apk --no-cache add \
    postgresql-client gnupg

WORKDIR /app

COPY ./config.yaml ./
COPY --from=builder /app/app /usr/bin/backup
RUN chmod +x /usr/bin/backup
RUN mkdir -p /backup
CMD ["backup"]
