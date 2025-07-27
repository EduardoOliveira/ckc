package enrichment

import (
	"context"
	"testing"

	"github.com/EduardoOliveira/ckc/types"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/assert"
)

func TestFetchIPData(t *testing.T) {
	enrichmentIP := EnrichmentIP{
		apiKey: "5abc58426eaed888866c98b19c20cf9134b6e1a0d1102fa2bfeff80a6c7f6f7f3243a6f5977de9f",
	}
	// Define a test IP address
	ip := types.NewIPAddress("187.174.238.116")

	actual, err := enrichmentIP.fetchIPData(context.Background(), ip)
	assert.NoError(t, err, "Expected no error fetching IP data")
	snaps.MatchSnapshot(t, actual)
}
