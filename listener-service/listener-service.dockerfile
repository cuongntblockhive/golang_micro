FROM golang:1.18-alpine as builder

WORKDIR /app

COPY . /app

RUN CGO_ENABLED=0 go build -o listenerApp .

RUN chmod +x /app/listenerApp

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/listenerApp .

CMD ["./listenerApp"]