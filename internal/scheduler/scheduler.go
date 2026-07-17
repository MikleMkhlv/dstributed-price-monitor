package scheduler

import (
	"context"
	"dstributed-price-monitor/internal/source"
	"fmt"
	"log"
	"time"
)

func RunScheduler(ctx context.Context, src *source.Source, tik time.Duration, tasksCh chan source.Record, errCh chan error) {
	if src == nil || src.Sources == nil {
		select {
		case errCh <- fmt.Errorf("source.UnidataFLSource.Pool: nil source"):
		default:
		}
		close(tasksCh)
		return
	}

	timer := time.NewTimer(tik)
	defer timer.Stop()
	defer close(tasksCh)

	for {
		select {
		case <-ctx.Done():
			select {
			case errCh <- fmt.Errorf("source.UnidataFLSource.Pool: context cancel"):
			default:
			}
			return
		case <-timer.C:
			log.Print("scheduler.RunScheduler: scheduler is run")
			for name, data := range src.Sources {
				log.Printf("scheduler.RunScheduler: source %s took on the task", name)
				select {
				case <-ctx.Done():
					select {
					case errCh <- fmt.Errorf("source.UnidataFLSource.Pool: context cancel"):
					default:
					}
					return
				case tasksCh <- data:
				}
			}
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			timer.Reset(tik)
		}
	}
}
