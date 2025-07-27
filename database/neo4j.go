package database

import (
	"context"
	"fmt"

	"github.com/EduardoOliveira/ckc/types"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// Neo4jConfig contains the configuration for connecting to a Neo4j database
type Neo4jConfig struct {
	URI      string
	Username string
	Password string
	Database string
}

// Neo4jClient provides access to the Neo4j database
type Neo4jClient struct {
	driver  neo4j.DriverWithContext
	config  Neo4jConfig
	context context.Context
}

// NewNeo4jClient creates a new Neo4j client with the given configuration
func NewNeo4jClient(ctx context.Context, config Neo4jConfig) (*Neo4jClient, error) {
	// Set default database if not provided
	if config.Database == "" {
		config.Database = "neo4j"
	}

	// Create authentication configuration
	auth := neo4j.BasicAuth(config.Username, config.Password, "")

	// Create Neo4j driver
	driver, err := neo4j.NewDriverWithContext(config.URI, auth)
	if err != nil {
		return nil, fmt.Errorf("failed to create Neo4j driver: %w", err)
	}

	// Verify connectivity
	err = driver.VerifyConnectivity(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Neo4j: %w", err)
	}

	return &Neo4jClient{
		driver:  driver,
		config:  config,
		context: ctx,
	}, nil
}

// Close closes the Neo4j driver
func (c *Neo4jClient) Close(ctx context.Context) error {
	return c.driver.Close(ctx)
}

// ExecuteQuery executes a query and returns the results
func (c *Neo4jClient) ExecuteQuery(
	ctx context.Context,
	cypher string,
	params map[string]any,
	db ...string,
) (neo4j.ResultWithContext, error) {
	// Use specified database or default to the configured one
	database := c.config.Database
	if len(db) > 0 && db[0] != "" {
		database = db[0]
	}

	// Create session
	session := c.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: database,
	})
	defer session.Close(ctx)

	// Execute query
	return session.Run(ctx, cypher, params)
}

func (c *Neo4jClient) Write(ctx context.Context, cyphers ...types.Cypher) (any, error) {
	cypherString := ""
	params := make(map[string]any)
	if len(cyphers) == 0 {
		return nil, fmt.Errorf("no Cypher queries provided")
	}
	for _, cypher := range cyphers {
		c, p := cypher.ToCypher() // Last query should return results
		if c == "" {
			cypherString += "\n"
			continue
		}
		cypherString = fmt.Sprintf("%s%s\n", cypherString, c)

		for k, v := range p {
			if _, exists := params[k]; exists {
				return nil, fmt.Errorf("duplicate parameter %s in Cypher query,", k)
			}
			params[k] = v
		}
	}
	cypherString = fmt.Sprintf("%s\nFINISH", cypherString)

	fmt.Println("Executing Cypher:", cypherString)

	// Execute write transaction
	return c.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		return tx.Run(ctx, cypherString, params)
	})
}

// ExecuteWrite executes a write transaction
func (c *Neo4jClient) ExecuteWrite(
	ctx context.Context,
	work func(tx neo4j.ManagedTransaction) (any, error),
	db ...string,
) (any, error) {
	// Use specified database or default to the configured one
	database := c.config.Database
	if len(db) > 0 && db[0] != "" {
		database = db[0]
	}

	// Create session
	session := c.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: database,
		AccessMode:   neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	// Execute transaction
	return session.ExecuteWrite(ctx, work)
}

// ExecuteRead executes a read transaction
func (c *Neo4jClient) ExecuteRead(
	ctx context.Context,
	work func(tx neo4j.ManagedTransaction) (any, error),
	db ...string,
) (any, error) {
	// Use specified database or default to the configured one
	database := c.config.Database
	if len(db) > 0 && db[0] != "" {
		database = db[0]
	}

	// Create session
	session := c.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: database,
		AccessMode:   neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	// Execute transaction
	return session.ExecuteRead(ctx, work)
}
