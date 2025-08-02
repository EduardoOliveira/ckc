package enrichment

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/EduardoOliveira/ckc/internal/opt"
	"github.com/EduardoOliveira/ckc/neo4j"
	"github.com/EduardoOliveira/ckc/types"
)

type AIPDBEnricher struct {
	ctx       context.Context
	apiKey    string
	neoClient *neo4j.Neo4jClient
}

func (_ *AIPDBEnricher) Name() string {
	return "AIPDB"
}

func NewAIPDBEnricher(ctx context.Context, apiKey string, n *neo4j.Neo4jClient) AIPDBEnricher {
	ensurePool(ctx)
	return AIPDBEnricher{
		ctx:       ctx,
		apiKey:    apiKey,
		neoClient: n,
	}
}

func (e *AIPDBEnricher) Enrich(parsed types.ParsedEvent) {
	// Check if the IP was enriched in the past 24 hours
	lastEnriched, err := e.neoClient.GetLastEnrichedAt(e.ctx, parsed.IPAddress, "AIPDBData")
	if err != nil {
		slog.WarnContext(e.ctx, "Failed to get last enrichment time, proceeding with enrichment", "ip", parsed.IPAddress, "error", err)
	} else {
		slog.InfoContext(e.ctx, "Last enrichment time", "ip", parsed.IPAddress, "last_enriched", lastEnriched)
		if !lastEnriched.IsZero() && time.Since(lastEnriched) < 24*time.Hour {
			slog.InfoContext(e.ctx, "IP was enriched less than 24 hours ago, skipping",
				"ip", parsed.IPAddress,
				"last_enriched", lastEnriched,
				"hours_ago", time.Since(lastEnriched).Hours())
			return
		}
	}
	slog.InfoContext(e.ctx, "Enriching IP with AIPDB", "ip", parsed.IPAddress)
	timeout, cancel := context.WithTimeout(e.ctx, 60*time.Second)
	defer cancel()

	select {
	case <-timeout.Done():
		slog.WarnContext(timeout, "Timeout while waiting for enrichment", "ip", parsed.IPAddress)
	case <-e.ctx.Done():
		slog.WarnContext(timeout, "Global context cancelled while waiting for enrichment", "ip", parsed.IPAddress)
	default:
		job, out := e.getData(timeout, parsed.IPAddress)
		publishJob(job)
		result := <-out
		if result.Error != nil {
			slog.ErrorContext(timeout, "Failed to enrich IP", "ip", parsed.IPAddress, "error", result.Error)
			return
		}
		slog.InfoContext(timeout, "Enrichment result", "ip", parsed.IPAddress)
		if err := e.neoClient.SaveAIPDBData(timeout, parsed.IPAddress, result.Value); err != nil {
			slog.ErrorContext(timeout, "Failed to save AIPDB data", "ip", parsed.IPAddress, "error", err)
			return
		}
	}
}

func (e *AIPDBEnricher) getData(ctx context.Context, ip types.IPAddress) (job, <-chan opt.Result[types.AIPDBData]) {
	done := make(chan opt.Result[types.AIPDBData], 1)
	return func() {
		req, err := http.NewRequestWithContext(ctx, "GET", "https://api.abuseipdb.com/api/v2/check?verbose=false&ipAddress="+ip.Address, nil)
		if err != nil {
			done <- opt.Err[types.AIPDBData](fmt.Errorf("failed to create request: %w", err))
			return
		}
		req.Header.Set("Key", e.apiKey)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			done <- opt.Err[types.AIPDBData](fmt.Errorf("failed to enrich IP %s: %w", ip.Address, err))
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			done <- opt.Err[types.AIPDBData](fmt.Errorf("failed to enrich IP %s: %s", ip.Address, resp.Status))
		}

		var response types.AbuseIPDBResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			done <- opt.Err[types.AIPDBData](fmt.Errorf("failed to decode response for IP %s: %w", ip.Address, err))
			return
		}
		done <- opt.Ok(response.Data)
	}, done
}
