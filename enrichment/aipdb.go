package enrichment

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
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
	if err := e.enrich(parsed.IPAddress); err != nil {
		slog.ErrorContext(e.ctx, "Failed to enrich IP with AIPDB", "ip", parsed.IPAddress, "error", err)
	}
}

func (e *AIPDBEnricher) enrich(ip types.IPAddress) error {
	// Check if the IP was enriched in the past 24 hours
	lastEnriched, err := e.neoClient.GetLastEnrichedAt(e.ctx, ip, "AIPDBData")
	if err != nil {
		slog.WarnContext(e.ctx, "Failed to get last enrichment time, proceeding with enrichment", "ip", ip, "error", err)
		return nil
	} else {
		if !lastEnriched.IsZero() && time.Since(lastEnriched) < 24*time.Hour {
			slog.InfoContext(e.ctx, "IP was enriched less than 24 hours ago, skipping",
				"ip", ip,
				"last_enriched", lastEnriched,
				"hours_ago", time.Since(lastEnriched).Hours())
			return nil
		}
	}
	slog.InfoContext(e.ctx, "Enriching IP with AIPDB", "ip", ip)
	timeout, cancel := context.WithTimeout(e.ctx, 60*time.Second)
	defer cancel()

	select {
	case <-timeout.Done():
		return fmt.Errorf("timeout while waiting for enrichment of IP %s", ip.Address)
	case <-e.ctx.Done():
		return fmt.Errorf("context cancelled while waiting for enrichment of IP %s: %w", ip.Address, e.ctx.Err())
	default:
		job, out := e.getData(timeout, ip)
		publishJob(job)
		result := <-out
		if result.Error != nil {
			return fmt.Errorf("failed to enrich IP %s: %w", ip.Address, result.Error)
		}
		slog.InfoContext(timeout, "Enrichment result", "ip", ip)
		if err := e.neoClient.SaveAIPDBData(timeout, ip, result.Value); err != nil {
			return fmt.Errorf("failed to save AIPDB data for IP %s: %w", ip.Address, err)
		}
	}
	return nil
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

func (e *AIPDBEnricher) EnrichAll(ctx context.Context) error {
	slog.InfoContext(ctx, "Starting AIPDB enrichment for all IPs")
	ips, err := e.neoClient.IterOverIPAddresses(ctx)
	if err != nil {
		return fmt.Errorf("failed to iterate over IP addresses: %w", err)
	}

	for ip, err := range ips {
		if err != nil {
			slog.WarnContext(ctx, "Failed to get IP address", "error", err)
			continue
		}
		e.enrich(ip)
	}

	slog.InfoContext(ctx, "Completed AIPDB enrichment for all IPs")
	return nil
}
