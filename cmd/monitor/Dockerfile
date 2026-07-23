FROM golang:1.25.0 AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o monitor ./cmd/monitor

FROM alpine:3.19
WORKDIR /app

COPY --from=builder /build/monitor .

EXPOSE 8081

CMD ["./monitor", "-configPath=conf_test.yaml"]
