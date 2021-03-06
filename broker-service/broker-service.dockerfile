FROM golang:1.18-alpine as builder

WORKDIR /app

COPY . /app

RUN CGO_ENABLED=0 go build -o brokerApp ./cmd/api

RUN chmod +x /app/brokerApp

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/brokerApp .

CMD ["./brokerApp"]

