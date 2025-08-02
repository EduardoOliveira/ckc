package enrichment

import (
	"context"
	"sync"
)

type job func()

var (
	workercount = 10
	jobs        = make(chan job, workercount)
	startOnce   = &sync.Once{}
)

func worker(ctx context.Context, jobs <-chan job) {
	for {
		select {
		case <-ctx.Done():
			return
		case j, ok := <-jobs:
			if ok {
				j()
			}
		}
	}
}

func publishJob(j job) {
	jobs <- j
}

func ensurePool(ctx context.Context) {
	startOnce.Do(func() {
		go func() {
			for range workercount {
				go worker(ctx, jobs)
			}
			defer close(jobs)
			<-ctx.Done()
		}()
	})
}
