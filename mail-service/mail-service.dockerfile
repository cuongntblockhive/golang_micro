FROM golang:1.18-alpine as builder

WORKDIR /app

COPY . /app

RUN CGO_ENABLED=0 go build -o mailApp ./cmd/api

RUN chmod +x /app/mailApp

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/mailApp ./
COPY template ./template
CMD ["/root/mailApp"]