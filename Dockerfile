FROM golang:1.22.5 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app main.go

FROM postgres:16-alpine

WORKDIR /app

COPY --from=builder /app/app /app
RUN chmod +x ./app
RUN mkdir -p /backup
CMD ["./app"]
