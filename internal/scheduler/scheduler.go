package scheduler

import (
	"context"
	"dstributed-price-monitor/internal/source"
	"log"
	"sync"
	"time"
)

func RunScheduler(ctx context.Context, src *source.Source, tik time.Duration, outCh chan source.ServiceData, errCh chan error) {
	t := time.NewTicker(tik)
	defer t.Stop()
	wg := sync.WaitGroup{}
	defer wg.Wait()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			log.Print("scheduler.RunScheduler: scheduler is run")
			for name, data := range src.Sources {
				wg.Add(1)
				log.Printf("scheduler.RunScheduler: siurce %s took on the task", name)
				go func(nameSource string, rec source.Record) {
					defer wg.Done()
					rec.Pool(ctx, outCh, errCh)
				}(name, data)
			}
			t.Reset(tik)
		}
	}
}
