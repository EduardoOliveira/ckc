package enrichment

import (
	"context"
	"math/rand"
	"testing"
	"time"
)

func TestWorkers(t *testing.T) {
	// This test is to ensure that the worker pool starts correctly
	// and that jobs can be published to it.
	ctx, cancel := context.WithTimeout(t.Context(), 1*time.Minute)
	defer cancel()

	ensurePool(ctx)

	for i := range 100 {
		publishJob(func(inner int) func() {
			return func() {
				time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond) // Simulate some work
				println("Processing job", inner)
			}
		}(i))
	}
	select {
	case <-ctx.Done():
		t.Error("context should not be done yet")
	case <-time.After(1 * time.Second):
	}
}
