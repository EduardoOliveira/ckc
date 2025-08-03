package neo4j

import (
	"context"
	"fmt"
	"iter"
	"log/slog"
	"time"

	"github.com/EduardoOliveira/ckc/types"
	n "github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func (c *Neo4jClient) IterOverIPAddresses(ctx context.Context) (iter.Seq2[types.IPAddress, error], error) {
	cypher := `
MATCH (ip:IPAddress)
RETURN ip.address AS address, ip.see AS seen, datetime(ip.last_seen) AS last_seen, datetime(ip.first_seen) AS first_seen`
	props := map[string]any{}
	result, err := c.ExecuteQuery2(ctx, cypher, props)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	return func(yeld func(types.IPAddress, error) bool) {
		for _, r := range result.Records {
			slog.InfoContext(ctx, "Processing record", "record", r)
			ip, err := mapRecordToIpAddress(r)
			if err != nil {
				_ = !yeld(types.IPAddress{}, fmt.Errorf("failed to map record to IP address: %w", err))
				return
			}
			if !yeld(ip, nil) {
				break
			}
		}
	}, nil
}

func mapRecordToIpAddress(r *n.Record) (types.IPAddress, error) {
	rtn := types.IPAddress{}
	address, isNil, err := n.GetRecordValue[string](r, "address")
	if err != nil {
		return types.IPAddress{}, fmt.Errorf("failed to get IP address: %w", err)
	}
	if !isNil {
		rtn.Address = address
	}
	seen, isNil, err := n.GetRecordValue[int64](r, "seen")
	if err != nil {
		return types.IPAddress{}, fmt.Errorf("failed to get seen count: %w", err)
	}
	if !isNil {
		rtn.Seen = seen
	}
	firstSeen, isNil, err := n.GetRecordValue[time.Time](r, "first_seen")
	if err != nil {
		return types.IPAddress{}, fmt.Errorf("failed to get first seen time: %w", err)
	}
	if !isNil {
		rtn.FirstSeen = firstSeen
	}
	lastSeen, isNil, err := n.GetRecordValue[time.Time](r, "last_seen")
	if err != nil {
		return types.IPAddress{}, fmt.Errorf("failed to get last seen time: %w", err)
	}
	if !isNil {
		rtn.LastSeen = lastSeen
	}
	return rtn, nil
}
