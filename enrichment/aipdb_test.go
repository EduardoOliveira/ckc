package enrichment

import (
	"context"
	"testing"

	"github.com/EduardoOliveira/ckc/internal/cfg"
	"github.com/EduardoOliveira/ckc/types"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/assert"
)

func TestFetchIPData(t *testing.T) {
	t.Parallel()

	t.Run("found", func(t *testing.T) {
		enrichmentIP := AIPDBEnricher{
			apiKey: cfg.Must("AIPDB_API_KEY"),
		}
		// Define a test IP address
		ip := types.IPAddress{Address: "187.174.238.116"}

		job, out := enrichmentIP.getData(context.Background(), ip)
		go job()
		result := <-out
		assert.NotNil(t, result)
		snaps.MatchSnapshot(t, result)
	})
}
