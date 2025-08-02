package neo4j

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/EduardoOliveira/ckc/types"
	neo "github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type neo4jSSHD struct {
	client *Neo4jClient
}

func (n *neo4jSSHD) Name() string {
	return "sshd_neo4j_store"
}

func NewNeo4jSSHD(client *Neo4jClient) neo4jSSHD {
	return neo4jSSHD{
		client: client,
	}
}

func (n *neo4jSSHD) Store(ctx context.Context, event types.ParsedEvent) error {
	slog.Info("Storing SSHD event in Neo4j", "event", event)
	res, err := n.client.ExecuteWrite(ctx, func(tx neo.ManagedTransaction) (any, error) {
		cypher, props := n.storeCypher(event)
		_, err := n.client.ExecuteQuery(ctx, cypher, props)
		if err != nil {
			return nil, fmt.Errorf("failed to execute Cypher query: %w", err)
		}
		return nil, nil
	})
	if err != nil {
		slog.Error("Failed to store SSHD event in Neo4j", "error", err)
		return fmt.Errorf("failed to store SSHD event in Neo4j: %w", err)
	}
	slog.Info("Successfully stored SSHD event in Neo4j", "result", res)

	return nil
}

func (n *neo4jSSHD) storeCypher(event types.ParsedEvent) (string, map[string]any) {
	cypher := `
		MERGE (s:Service {name: $serviceName, port: $port, host: $host})
		ON CREATE SET s.first_seen = datetime($ingestion), s.seen = 0
		SET s.seen = s.seen + 1
		WITH *

		MERGE (ip:IPAddress {address: $ip_address})
		ON CREATE SET ip.seen = 0, ip.first_seen = datetime($ingestion)
		SET ip.last_seen = datetime($ingestion)
		SET ip.seen = ip.seen + 1
		WITH *

		MERGE (ip)-[ct:CONNECTED_TO]->(s)
		ON CREATE SET ct.times = 0, ct.fist_time = datetime($ingestion)
		SET ct.times = ct.times + 1, ct.last_time = datetime($ingestion)
		WITH *
 
		MERGE (username:Username {name: $username})
		ON CREATE SET username.first_seen = datetime($ingestion), username.seen = 0
		SET username.last_seen = datetime($ingestion), username.seen = username.seen + 1
		WITH *

		MERGE (ip)-[w:WITH_USERNAME]->(username)
		ON CREATE SET w.first_time = datetime($ingestion), w.times = 0
		SET w.last_time = datetime($ingestion), w.times = w.times + 1
		WITH *

		MERGE (username)-[a:AUTHENTICATED_ON]->(s)
		ON CREATE SET a.first_time = datetime($ingestion), a.failures = 0, a.successes = 0, a.times = 0
		SET a.last_time = datetime($ingestion), a.times = a.times + 1,
	`
	if event.SSHDEvent.OrElse(types.SSHDParsedEvent{}).Success {
		cypher += `a.successes = a.successes + 1
		`
	} else {
		cypher += `a.failures = a.failures + 1
		`
	}

	cypher = fmt.Sprintf("%s\nFINISH", cypher)

	params := map[string]any{
		"serviceName": event.Service.Name,
		"port":        event.Service.Port,
		"host":        event.Service.Host,
		"ip_address":  event.IPAddress.Address,
		"ingestion":   event.Ingestion.Format(time.RFC3339),
		"username":    event.Username.Name,
	}

	return cypher, params
}
