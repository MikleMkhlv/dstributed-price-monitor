# dstributed-price-monitor

1. APPLICATION_ADDRESS="127.0.0.1" APPLICATION_PORT="8081" go run --race cmd/monitor/main.go -configPath=conf_test.yaml

2. APPLICATION_ADDRESS="127.0.0.1" APPLICATION_PORT="8080" go run -race ./cmd/fetcher -configPath=conf_test.yaml