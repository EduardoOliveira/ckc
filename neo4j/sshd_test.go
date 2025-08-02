package neo4j

import (
	"testing"

	"github.com/EduardoOliveira/ckc/internal/time_help"
	"github.com/EduardoOliveira/ckc/types"
	"github.com/gkampitakis/go-snaps/snaps"
)

func TestSSDHStoreCypher(t *testing.T) {
	h := neo4jSSHD{}
	cypher, props := h.storeCypher(types.ParsedEvent{
		Ingestion:   time_help.Now(),
		ServiceName: "sshd",
		IPAddress:   types.IPAddress{Address: "127.0.0.1"},
		Username:    types.Username{Name: "root"},
		Service: types.Service{
			Name: "sshd",
			Port: 22,
			Host: "localhost",
		},
	})
	snaps.MatchSnapshot(t, cypher, props)
}
