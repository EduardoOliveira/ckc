package neo4j

import (
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

func TestMergeCountryCypher(t *testing.T) {
	t.Run("test simple", func(t *testing.T) {
		t.Parallel()
		key, cypher, props := MergeCountryCypher("pt", "Portugal", map[string]any{
			"country_code":  "PT",
			"has_coastline": true,
		})
		snaps.MatchSnapshot(t, key, cypher, props)
	})
}

func TestMergeCityWithCountryCypher(t *testing.T) {
	t.Run("test simple", func(t *testing.T) {
		t.Parallel()
		key, cypher, props := MergeCityWithCountryCypher(1, "Lisbon", "Portugal", map[string]any{
			"population": 500000,
			"area":       100,
		})
		snaps.MatchSnapshot(t, key, cypher, props)
	})
}

func TestMatchIPAddressCypher(t *testing.T) {
	t.Run("test simple", func(t *testing.T) {
		t.Parallel()
		key, cypher := MergeIPAddressCypher(1, "127.0.0.1")
		snaps.MatchSnapshot(t, key, cypher)
	})
}
