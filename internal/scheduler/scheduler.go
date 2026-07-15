package scheduler

import (
	"context"
	"dstributed-price-monitor/internal/source"
	"fmt"
	"log"
	"time"
)

func RunScheduler(ctx context.Context, src *source.Source, tik time.Duration, tasksCh chan source.Record, errCh chan error) {
	t := time.NewTicker(tik)
	defer t.Stop()
	defer close(tasksCh)
	for {
		select {
		case <-ctx.Done():
			errCh <- fmt.Errorf("source.UnidataFLSource.Pool: context cancel")
			return
		case <-t.C:
			log.Print("scheduler.RunScheduler: scheduler is run")
			for name, data := range src.Sources {
				log.Printf("scheduler.RunScheduler: siurce %s took on the task", name)
				select {
				case <-ctx.Done():
					errCh <- fmt.Errorf("source.UnidataFLSource.Pool: context cancel")
					return
				case tasksCh <- data:
				}
			}
			t.Reset(tik)
		}
	}
}
