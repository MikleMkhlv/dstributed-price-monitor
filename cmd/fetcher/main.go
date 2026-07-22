package main

import (
	"context"
	dto "dstributed-price-monitor/api/dto"
	fet "dstributed-price-monitor/internal/fetcher"
	"log"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	fetchCh := make(chan dto.FetchRequest, 50)
	cfg := fet.NewFetchConfig()
	server := fet.NewServer(fetchCh, cfg)

	go func() {
		if err := server.RunServer(); err != nil {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()
	go func() {
		close(fetchCh)
		server.Stop()
	}()
}
