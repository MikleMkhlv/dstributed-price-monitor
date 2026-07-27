package main

import (
	"context"
	"dstributed-price-monitor/api/dto"
	"dstributed-price-monitor/config"
	"dstributed-price-monitor/internal/broker"
	"dstributed-price-monitor/internal/fetcher/mapper"
	"dstributed-price-monitor/internal/logs"
	"dstributed-price-monitor/internal/source"
	"dstributed-price-monitor/internal/worker"
	"flag"
	"fmt"
	"log"
	"os"
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
	// cfgFtc := fet.NewFetchConfig()
	cfg := config.MustLoadConfig(configPath())
	// server := fet.NewServer(fetchCh, cfgFtc)
	// client := fet.NewClient(outCh, cfg)
	worker := worker.New(cfg.Scheduler.CountWorker, cfg.Scheduler.MaxCalls)
	nutsConn, err := broker.NewNats(cfg)
	if err != nil {
		log.Fatal(err)
	}
	sub, err := broker.NewSubscription(nutsConn, fetchCh)
	pub, err := broker.NewPublisher[dto.FetchResponce](nutsConn)
	fechMap := mapper.FetchMaper{}

	// wg.Add(1)
	// go func() {
	// 	defer wg.Done()
	// 	if err := server.RunServer(); err != nil {
	// 		log.Fatal(err)
	// 	}
	// }()

	// wg.Add(1)
	// go func() {
	// 	defer wg.Done()
	// 	client.SendToMonitor(ctx)
	// }()
	if err := sub.Start(cfg.Nats.Queues.InFetch, "fromMonitor"); err != nil {
		log.Print(err)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		worker.RunWorker(ctx, fetchCh, outCh, errorCh)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		path := "/data/logs/fetcher_errors.log"
		info, err := os.Stat(path)
		if err == nil && info.IsDir() {
			fmt.Printf("Fatal error: path %s occupied by the directory !\n", path)
			return
		}
		file, err := os.Create(path)
		if err != nil {
			log.Fatalf("error creete log file from errors. %v", err)
		}
		defer file.Close()
		logs.LogWriter(file, errorCh)
	}()

	go func() {
		for data := range outCh {
			resp, err := fechMap.CitizenToFetchResponse(data)
			log.Printf("feature.main: resp: %v", resp)
			if err != nil {
				log.Print(err)
			}
			pub.Publish(cfg.Nats.Queues.InMonitor, *resp)
		}
	}()

	<-ctx.Done()
	go func() {
		close(fetchCh)
		close(errorCh)
		close(outCh)
		sub.Stop()
		nutsConn.Close()
		// server.Stop()
	}()
}

func configPath() string {
	filePath := flag.String("configPath", "", "configuration file")
	flag.Parse()
	return *filePath
}
