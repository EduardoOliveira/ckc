package enrichment

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"os"

	"github.com/EduardoOliveira/ckc/database"
	"github.com/EduardoOliveira/ckc/types"
)

type EnrichmentIP struct {
	neo4jClient *database.Neo4jClient
	apiKey      string
}

func NewEnrichmentIP(neo4jClient *database.Neo4jClient) *EnrichmentIP {
	var key string
	if key = os.Getenv("API_KEY"); key != "" {
		panic("API_KEY environment variable is not set")
	}

	return &EnrichmentIP{
		neo4jClient: neo4jClient,
		apiKey:      key,
	}
}

func (e *EnrichmentIP) Enrich(ctx context.Context, ip types.IPAddress) error {
	go func() {
		// check last fetched time
		/*
				lastFetched, err := e.neo4jClient.GetEnrichmentLastFetched(ctx, ip.Address)
				if err != nil {
					slog.Error("Failed to get last fetched time", "ip", ip, "error", err)
					return
				}

				if time.Since(lastFetched) < 24*time.Hour {
			slog.Info("IP data is recent, skipping enrichment", "ip", ip, "last_fetched", lastFetched)
		*/
		data, err := e.fetchIPData(ctx, ip)
		slog.Info("Fetched IP data", "ip", ip, "data", data)
		if err != nil {
			slog.Error("Failed to fetch IP data", "ip", ip, "error", err)
			return
		}
		_, err = e.neo4jClient.Write(ctx, &data)
		if err != nil {
			slog.Error("Failed to write IP data", "ip", ip, "error", err)
			return
		}
		//	}
	}()
	return nil
}

func (e *EnrichmentIP) fetchIPData(ctx context.Context, ip types.IPAddress) (types.AIPDBData, error) {
	// fetch new data from AbuseIPDB
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.abuseipdb.com/api/v2/check?verbose=false&ipAddress="+ip.Address, nil)
	if err != nil {
		return types.AIPDBData{}, err
	}
	req.Header.Set("Key", e.apiKey)

	b, _ := httputil.DumpRequest(req, false)
	slog.Debug("Request to AbuseIPDB", "request", string(b))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return types.AIPDBData{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Error("Failed to enrich IP", "ip", ip, "status", resp.Status)
		return types.AIPDBData{}, fmt.Errorf("failed to enrich IP %s: %s", ip.Address, resp.Status)
	}

	var response types.AbuseIPDBResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		slog.Error("Failed to decode response", "ip", ip, "error", err)
		return types.AIPDBData{}, err
	}
	slog.Info("Enriched IP", "ip", ip, "data", response.Data)
	return response.Data, nil
}
