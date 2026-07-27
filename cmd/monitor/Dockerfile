FROM golang:1.25.0 AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o monitor ./cmd/monitor

FROM alpine:3.19
WORKDIR /app
RUN mkdir -p /data/logs && chmod 777 /data/logs

COPY --from=builder /build/monitor .

EXPOSE 8081 5432 5432/tcp 6379 6379/tcp 6379/udp

CMD ["./monitor", "-configPath=conf_test.yaml"]
