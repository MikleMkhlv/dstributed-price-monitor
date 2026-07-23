package main

import (
	"context"
	"dstributed-price-monitor/config"
	fet "dstributed-price-monitor/internal/fetcher"
	"dstributed-price-monitor/internal/source"
	"dstributed-price-monitor/internal/worker"
	"flag"
	"log"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	wg := sync.WaitGroup{}
	fetchCh := make(chan source.Record, 50)
	errorCh := make(chan error, 100)
	outCh := make(chan source.ServiceData)
	cfgFtc := fet.NewFetchConfig()
	cfg := config.MustLoadConfig(configPath())
	server := fet.NewServer(fetchCh, cfgFtc)
	worker := worker.New(cfg.Scheduler.CountWorker, cfg.Scheduler.MaxCalls)

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := server.RunServer(); err != nil {
			log.Fatal(err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		worker.RunWorker(ctx, fetchCh, outCh, errorCh)
	}()

	<-ctx.Done()
	go func() {
		close(fetchCh)
		close(errorCh)
		close(outCh)
		server.Stop()
	}()
}

func configPath() string {
	filePath := flag.String("configPath", "", "configuration file")
	flag.Parse()
	return *filePath
}
