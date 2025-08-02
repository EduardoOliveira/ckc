package neo4j

import (
	"context"
	"errors"
	"fmt"

	"github.com/EduardoOliveira/ckc/types"
)

func (n *Neo4jClient) GetUsername(ctx context.Context, username string) (types.Username, error) {
	query := `
		MATCH (u:Username {name: $username})
		RETURN u.name AS name, u.first_seen AS first_seen, u.last_seen AS last_seen, u.seen AS seen
	`
	params := map[string]any{
		"username": username,
	}

	result, err := n.ExecuteQuery(ctx, query, params)
	if err != nil {
		return types.Username{}, err
	}
	record, err := result.Single(ctx)
	if err != nil {
		if err.Error() == "no more records" {
			return types.Username{}, errors.New("Not found") // Username not found
		}
		return types.Username{}, fmt.Errorf("failed to get username: %w", err)
	}

	return types.MapUsernameFromMap(record.AsMap())
}
