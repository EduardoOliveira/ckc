package neo4j

import (
	"context"
	"fmt"
	"time"

	n "github.com/neo4j/neo4j-go-driver/v5/neo4j"
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
	driver  n.DriverWithContext
	config  Neo4jConfig
	context context.Context
	now     func() time.Time
}

// NewNeo4jClient creates a new Neo4j client with the given configuration
func NewNeo4jClient(ctx context.Context, config Neo4jConfig) (*Neo4jClient, error) {
	// Set default database if not provided
	if config.Database == "" {
		config.Database = "neo4j"
	}

	// Create authentication configuration
	auth := n.BasicAuth(config.Username, config.Password, "")

	// Create Neo4j driver
	driver, err := n.NewDriverWithContext(config.URI, auth)
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
		now:     time.Now,
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
) (n.ResultWithContext, error) {
	// Use specified database or default to the configured one
	database := c.config.Database
	if len(db) > 0 && db[0] != "" {
		database = db[0]
	}

	// Create session
	session := c.driver.NewSession(ctx, n.SessionConfig{
		DatabaseName: database,
	})
	defer session.Close(ctx)

	// Execute query
	return session.Run(ctx, cypher, params)
}

// ExecuteWrite executes a write transaction
func (c *Neo4jClient) ExecuteWrite(
	ctx context.Context,
	work func(tx n.ManagedTransaction) (any, error),
	db ...string,
) (any, error) {
	// Use specified database or default to the configured one
	database := c.config.Database
	if len(db) > 0 && db[0] != "" {
		database = db[0]
	}

	// Create session
	session := c.driver.NewSession(ctx, n.SessionConfig{
		DatabaseName: database,
		AccessMode:   n.AccessModeWrite,
	})
	defer session.Close(ctx)

	// Execute transaction
	return session.ExecuteWrite(ctx, work)
}

// ExecuteRead executes a read transaction
func (c *Neo4jClient) ExecuteRead(
	ctx context.Context,
	work func(tx n.ManagedTransaction) (any, error),
	db ...string,
) (any, error) {
	// Use specified database or default to the configured one
	database := c.config.Database
	if len(db) > 0 && db[0] != "" {
		database = db[0]
	}

	// Create session
	session := c.driver.NewSession(ctx, n.SessionConfig{
		DatabaseName: database,
		AccessMode:   n.AccessModeRead,
	})
	defer session.Close(ctx)

	// Execute transaction
	return session.ExecuteRead(ctx, work)
}
