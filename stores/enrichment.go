package stores

import (
	"context"
	"time"
)

func (c *Neo4jClient) GetEnrichmentLastFetched(ctx context.Context, address string) (time.Time, error) {
	query := `MATCH (en:Enrichment {address: $address})
	RETURN en.last_fetched AS last_fetched`
	params := map[string]any{
		"address": address,
	}

	result, err := c.ExecuteQuery(ctx, query, params)
	if err != nil {
		return time.Time{}, err
	}
	if result.Next(ctx) {
		lastFetched, ok := result.Record().Get("last_fetched")
		if !ok {
			return time.Time{}, nil // No lastFetched found
		}
		if t, ok := lastFetched.(time.Time); ok {
			return t, nil
		}
	}
	return time.Time{}, nil // No lastFetched found or type mismatch
}
