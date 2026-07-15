package worker

import (
	"context"
	"dstributed-price-monitor/internal/source"
	"dstributed-price-monitor/internal/worker/semaphore"
	"fmt"
	"log"
	"sync"

	"github.com/google/uuid"
)

type Worker struct {
	count     int
	wg        sync.WaitGroup
	semaphore semaphore.Semaphore
}

func New(countWorker int, maxCallsSem int) *Worker {
	return &Worker{
		count:     countWorker,
		wg:        sync.WaitGroup{},
		semaphore: *semaphore.NewSemaphore(maxCallsSem),
	}
}

func (w *Worker) RunWorker(ctx context.Context, tasksCh chan source.Record, outCh chan source.ServiceData, errCh chan error) {
	for i := 1; i <= w.count; i++ {
		if !w.semaphore.AcquireCtx(ctx) {
			return
		}
		w.wg.Add(1)
		go func(workerIndex int) {
			defer w.semaphore.Release()
			ctxWithOper := context.WithValue(ctx, "operId", uuid.New())
			defer w.wg.Done()
			for task := range tasksCh {
				log.Printf("worker.RunWorker{%v}: worker-%d took on the task", ctxWithOper.Value("operId"), workerIndex)
				select {
				case <-ctxWithOper.Done():
					errCh <- fmt.Errorf("source.UnidataFLSource.Pool: context cancel")
					return
				default:
				}
				task.Pool(ctxWithOper, outCh, errCh)
			}
		}(i)
	}
}
