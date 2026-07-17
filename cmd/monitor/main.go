package main

import (
	"context"
	"dstributed-price-monitor/config"
	"dstributed-price-monitor/internal/agregator"
	"dstributed-price-monitor/internal/repository"
	"dstributed-price-monitor/internal/scheduler"
	"dstributed-price-monitor/internal/source"
	"dstributed-price-monitor/internal/worker"
	"flag"
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
	// TODO: errorCh Сейчас буфферизированный. Когда буфер заполнится, то программа залочится. В будущем надо канал этот читать
	errorCh := make(chan error, 100)
	tasksCh := make(chan source.Record)
	cfg := config.MustLoadConfig(configPath())
	sources := source.NewSource(*cfg)
	tic := time.Second * time.Duration(cfg.Scheduler.Timeout)
	worker := worker.New(cfg.Scheduler.CountWorker, cfg.Scheduler.MaxCalls)

	db := repository.NewPG(ctx, cfg)
	rds := repository.NewRedis(ctx, cfg)
	agr := agregator.New(db, rds)

	wg.Add(1)
	go func() {
		defer wg.Done()
		worker.RunWorker(ctx, tasksCh, outCh, errorCh)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		scheduler.RunScheduler(ctx, sources, tic, tasksCh, errorCh)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		agr.Comparable(ctx, outCh)
	}()

	<-ctx.Done()
	go func() {
		wg.Wait()
		close(outCh)
		close(errorCh)
		db.Close()
		rds.Close()
	}()
}

func configPath() string {
	filePath := flag.String("configPath", "", "configuration file")
	flag.Parse()
	return *filePath
}
