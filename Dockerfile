FROM golang:1.22.5 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o app main.go

FROM postgres:16

RUN apt-get update && apt-get install -y cron

COPY --from=builder /app/app /usr/local/bin/backup
RUN mkdir -p /backup
WORKDIR /backup
ENTRYPOINT ["backup"]
