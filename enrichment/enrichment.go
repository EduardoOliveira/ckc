package enrichment

import (
	"context"
	"time"

	"github.com/EduardoOliveira/ckc/database"
	"github.com/EduardoOliveira/ckc/types"
)

type Enrichment struct {
	ctx context.Context
	IPs chan (types.IPAddress)
}

func New(ctx context.Context, _ *database.Neo4jClient) *Enrichment {
	return &Enrichment{
		ctx: ctx,
		IPs: make(chan types.IPAddress, 10),
	}
}

func (e *Enrichment) Start() error {
	go func() {
		for {
			select {
			case <-e.ctx.Done():
				return
			case _ = <-e.IPs:
				// Simulate enrichment process
				time.Sleep(100 * time.Millisecond) // Simulating processing time
			}
		}
	}()

	return nil
}
