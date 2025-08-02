package neo4j

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/EduardoOliveira/ckc/types"
	n "github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func (c *Neo4jClient) SaveAIPDBData(ctx context.Context, target types.IPAddress, enrichment types.AIPDBData) error {
	cypher, props := c.saveAIPDBCypher(target, enrichment)

	slog.InfoContext(ctx, "Saving AIPDB data", "ip", target.Address)
	result, err := c.ExecuteWrite(ctx, func(tx n.ManagedTransaction) (any, error) {
		return tx.Run(ctx, cypher, props)
	})
	if err != nil {
		slog.ErrorContext(ctx, "Failed to save AIPDB data", "ip", target.Address, "result", result, "error", err, "cypher", cypher, "props", props)
		return fmt.Errorf("failed to save AIPDB data: %w", err)
	}

	return nil
}

func (c *Neo4jClient) saveAIPDBCypher(target types.IPAddress, enrichment types.AIPDBData) (string, map[string]any) {
	props := make(map[string]any)
	var cypher string
	cypher += `
MERGE (ip_1:IPAddress {address: $ip_address})
WITH ip_1`
	props["ip_address"] = target.Address

	cypher += `
MERGE (aipdb:AIPDBData {address: $aipdb_ip_address})
	SET aipdb.isp = $aipdb_isp, 
	aipdb.is_tor = $aipdb_is_tor, 
	aipdb.is_public = $aipdb_is_public,
	aipdb.is_whitelisted = $aipdb_is_whitelisted
WITH ip_1, aipdb 

MERGE (ip_1)-[enriched:ENRICHED_BY]->(aipdb)
SET enriched.last_enrichment = datetime($now)

	`
	props["aipdb_ip_address"] = target.Address
	props["aipdb_isp"] = enrichment.Isp
	props["aipdb_is_tor"] = enrichment.IsTor
	props["aipdb_is_public"] = enrichment.IsPublic
	props["aipdb_is_whitelisted"] = enrichment.IsWhitelisted
	props["now"] = c.now().Format(time.RFC3339)

	lowerContryCode := strings.ToLower(enrichment.CountryCode)
	cypher += fmt.Sprintf(`
MERGE (c:Country {country_code: $loc_country_code, country_name: $loc_country_name})
WITH ip_1, c
MERGE (ip_1)-[:LOCATED_IN]->(c)
WITH ip_1
	`)
	props["loc_country_code"] = lowerContryCode
	props["loc_country_name"] = enrichment.CountryName

	reportsCount := make(map[types.Country]int)
	for _, report := range enrichment.Reports {
		c := types.Country{
			Code: strings.ToLower(report.ReporterCountryCode),
			Name: report.ReporterCountryName,
		}
		if _, exists := reportsCount[c]; !exists {
			reportsCount[c] = 0
		}
		reportsCount[c]++
	}

	for c, count := range reportsCount {
		cypher += fmt.Sprintf(`
MERGE (c_%s:Country {country_code: $%s_r_country_code, country_name: $%s_r_country_name})
WITH ip_1, c_%s
MERGE (ip_1)-[:REPORTED_IN]->(c_%s)
	SET c_%s.times = %d 
WITH ip_1
`,
			c.Code,
			c.Code,
			c.Code,
			c.Code,
			c.Code,
			c.Code,
			count,
		)
		props[fmt.Sprintf("%s_r_country_code", c.Code)] = c.Code
		props[fmt.Sprintf("%s_r_country_name", c.Code)] = c.Name

	}
	cypher += "\nFINISH\n"
	return cypher, props
}

func (c *Neo4jClient) GetLastEnrichedAt(ctx context.Context, target types.IPAddress, enrichmentType string) (time.Time, error) {
	cypher := fmt.Sprintf(`
MATCH (ip:IPAddress)-[e:ENRICHED_BY]->(:%s)
WHERE ip.address = "%s"
RETURN e.last_enrichment AS last_enrichment
LIMIT 1
`, enrichmentType, target.Address)
	props := map[string]any{}
	res, err := c.ExecuteQuery2(ctx, cypher, props)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get last enriched at", "ip", target.Address, "error", err, "cypher", cypher, "props", props)
		return time.Time{}, fmt.Errorf("failed to get last enriched at: %w", err)
	}
	slog.InfoContext(ctx, "Query executed", "cypher", cypher, "props", props, "result", res)

	ts, _, err := n.GetRecordValue[time.Time](res.Records[0], "last_enrichment")
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get last enrichment time", "ip", target.Address, "error", err)
		return time.Time{}, fmt.Errorf("failed to get last enrichment time: %w", err)
	}

	return ts, nil
}
