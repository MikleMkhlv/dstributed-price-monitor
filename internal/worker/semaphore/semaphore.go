package semaphore

import "context"

type Semaphore struct {
	C chan struct{}
}

func NewSemaphore(capacityCh int) *Semaphore {
	return &Semaphore{
		C: make(chan struct{}, capacityCh),
	}
}

func (s *Semaphore) AcquireCtx(ctx context.Context) bool {
	select {
	case s.C <- struct{}{}:
		return true
	case <-ctx.Done():
		return false
	}
}

func (s *Semaphore) Release() {
	<-s.C
}
