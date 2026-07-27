package scheduler

import (
	"context"
	"dstributed-price-monitor/api/dto"
	"dstributed-price-monitor/config"
	"dstributed-price-monitor/internal/broker"
	"dstributed-price-monitor/internal/scheduler/mapper"
	"dstributed-price-monitor/internal/source"
	"fmt"
	"log"
	"time"
)

func RunScheduler(ctx context.Context, src *source.Source, tik time.Duration,
	publisher broker.Publisher[dto.FetchRequest], cfg *config.Config,
	tasksCh chan source.Record, errCh chan error,
) {
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

	mapper := mapper.MapperShd{}

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
				// case tasksCh <- data:
				default:
					ev := mapper.RecordToRequestForFether(data)

					if err := publisher.Publish(cfg.Nats.Queues.InFetch, ev); err != nil {
						select {
						case <-ctx.Done():
							select {
							case errCh <- fmt.Errorf("source.UnidataFLSource.Pool: context cancel"):
							default:
							}
						case errCh <- err:
						}
					}
					log.Printf("source.UnidataFLSource.Pool: send in queue {%s} is success", cfg.Nats.Queues.InFetch)
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
