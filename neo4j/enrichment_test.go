package neo4j

import (
	"testing"

	"github.com/EduardoOliveira/ckc/internal/time_help"
	"github.com/EduardoOliveira/ckc/types"
	"github.com/gkampitakis/go-snaps/snaps"
)

func TestNeo4jEnrichment(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		t.Parallel()
		c := &Neo4jClient{}
		c.now = time_help.Now
		cypher, params := c.saveAIPDBCypher(types.IPAddress{Address: "127.0.0.1"},
			types.AIPDBData{
				CountryName:   "Portugal",
				CountryCode:   "PT",
				IsPublic:      true,
				IsWhitelisted: false,
				IsTor:         false,
				Isp:           "ISP Example",
				Reports: []types.AIDBReport{
					{
						ReporterCountryName: "Spain",
						ReporterCountryCode: "ES",
					},
					{
						ReporterCountryName: "France",
						ReporterCountryCode: "FR",
					},
					{
						ReporterCountryName: "France",
						ReporterCountryCode: "FR",
					},
				},
			},
		)
		snaps.MatchSnapshot(t, cypher, params)
	})
	t.Run("reported in origin", func(t *testing.T) {
		t.Parallel()
		c := &Neo4jClient{}
		c.now = time_help.Now
		cypher, params := c.saveAIPDBCypher(types.IPAddress{Address: "127.0.0.1"},
			types.AIPDBData{
				CountryName:   "Portugal",
				CountryCode:   "PT",
				IsPublic:      true,
				IsWhitelisted: false,
				IsTor:         false,
				Isp:           "ISP Example",
				Reports: []types.AIDBReport{
					{
						ReporterCountryName: "Portugal",
						ReporterCountryCode: "PT",
					},
					{
						ReporterCountryName: "Spain",
						ReporterCountryCode: "ES",
					},
					{
						ReporterCountryName: "France",
						ReporterCountryCode: "FR",
					},
					{
						ReporterCountryName: "France",
						ReporterCountryCode: "FR",
					},
				},
			},
		)
		snaps.MatchSnapshot(t, cypher, params)
	})
}
