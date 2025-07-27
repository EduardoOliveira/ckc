package types

import "time"

type AbuseIPDBResponse struct {
	Data AIPDBData `json:"data"`
}
type AIDBReport struct {
	ReportedAt          time.Time `json:"reportedAt"`
	Comment             string    `json:"comment"`
	Categories          []int     `json:"categories"`
	ReporterID          int       `json:"reporterId"`
	ReporterCountryCode string    `json:"reporterCountryCode"`
	ReporterCountryName string    `json:"reporterCountryName"`
}
type AIPDBData struct {
	IPAddress            string       `json:"ipAddress"`
	IsPublic             bool         `json:"isPublic"`
	IPVersion            int          `json:"ipVersion"`
	IsWhitelisted        bool         `json:"isWhitelisted"`
	AbuseConfidenceScore int          `json:"abuseConfidenceScore"`
	CountryCode          string       `json:"countryCode"`
	CountryName          string       `json:"countryName"`
	UsageType            string       `json:"usageType"`
	Isp                  string       `json:"isp"`
	Domain               string       `json:"domain"`
	Hostnames            []any        `json:"hostnames"`
	IsTor                bool         `json:"isTor"`
	TotalReports         int          `json:"totalReports"`
	NumDistinctUsers     int          `json:"numDistinctUsers"`
	LastReportedAt       time.Time    `json:"lastReportedAt"`
	Reports              []AIDBReport `json:"reports"`

	// internal properties
	LastFetched time.Time `json:"lastFetched"`
}

func (data *AIPDBData) ToCypher() (cypher string, params map[string]any) {
	cypher = `MERGE (ipdb:Enrichment:AbuseIPDB{address: $ipdb_address})
	SET ipdb.is_public: $ipdb_is_public,
		ipdb.is_whitelisted: $ipdb_is_whitelisted,
		ipdb.ip_version: $ipdb_ip_version,
		ipdb.abuse_confidence_score: $ipdb_abuse_confidence_score,
		ipdb.country_code: $ipdb_country_code,
		ipdb.country_name: $ipdb_country_name,
		ipdb.usage_type: $ipdb_usage_type,
		ipdb.isp: $ipdb_isp,
		ipdb.domain: $ipdb_domain,
		ipdb.hostnames: $ipdb_hostnames,
		ipdb.is_tor: $ipdb_is_tor,
		ipdb.total_reports: $ipdb_total_reports,
		ipdb.num_distinct_users: $ipdb_num_distinct_users,
		ipdb.last_reported_at: $ipdb_last_reported_at,
		ipdb.reports: $ipdb_reports,
		ipdb.last_fetched: $ipdb_last_fetched
	WITH ipdb`

	params = map[string]any{
		"ipdb_address":                data.IPAddress,
		"ipdb_is_public":              data.IsPublic,
		"ipdb_is_whitelisted":         data.IsWhitelisted,
		"ipdb_ip_version":             data.IPVersion,
		"ipdb_abuse_confidence_score": data.AbuseConfidenceScore,
		"ipdb_country_code":           data.CountryCode,
		"ipdb_country_name":           data.CountryName,
		"ipdb_usage_type":             data.UsageType,
		"ipdb_isp":                    data.Isp,
		"ipdb_domain":                 data.Domain,
		"ipdb_hostnames":              data.Hostnames,
		"ipdb_is_tor":                 data.IsTor,
		"ipdb_total_reports":          data.TotalReports,
		"ipdb_num_distinct_users":     data.NumDistinctUsers,
		"ipdb_last_reported_at":       data.LastReportedAt,
		"ipdb_reports":                data.Reports,
		"ipdb_last_fetched":           data.LastFetched,
	}

	return
}
