FROM golang:1.18-alpine as builder

WORKDIR /app

COPY . /app

RUN CGO_ENABLED=0 go build -o frontApp ./cmd/web

RUN chmod +x /app/frontApp

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/frontApp .

CMD ["./frontApp"]