package main

import (
	"context"
	"dstributed-price-monitor/cmd/monitor/internal/scheduler"
	"dstributed-price-monitor/cmd/monitor/internal/scheduler/source"
	"dstributed-price-monitor/config"
	"flag"
	"log"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	wg := sync.WaitGroup{}
	outCh := make(chan source.ServiceData)
	errorCh := make(chan error, 10)
	cfg := config.MustLoadConfig(configPath())
	sources := source.NewSource(*cfg)

	tic := time.Second * time.Duration(cfg.Scheduler.Timeout)

	wg.Add(1)
	go func() {
		defer wg.Done()
		scheduler.RunScheduler(ctx, sources, tic, outCh, errorCh)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for data := range outCh {
			switch d := data.(type) {
			case source.Citizen:
				log.Printf("main.Consumer. Citizen:%s", d.String())
			case source.Organization:
				log.Printf("main.Consumer. Org:%s", d.String())
			default:
				log.Println("unknown type")
				continue
			}
		}
	}()

	<-ctx.Done()
	go func() {
		wg.Wait()
		close(outCh)
		close(errorCh)
	}()
}

func configPath() string {
	filePath := flag.String("configPath", "", "configuration file")
	flag.Parse()
	return *filePath
}
