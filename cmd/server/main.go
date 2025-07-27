package main

import (
	"context"
	"log"
	"os"

	"github.com/EduardoOliveira/ckc/database"
	"github.com/EduardoOliveira/ckc/enrichment"
	"github.com/EduardoOliveira/ckc/handler"
	syslog "gopkg.in/mcuadros/go-syslog.v2"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize Neo4j connection
	neo4jClient, err := setupNeo4jConnection(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to Neo4j: %v", err)
	}
	defer neo4jClient.Close(ctx)

	enricher := enrichment.New(ctx, neo4jClient)
	/*
		if err := enricher.Start(); err != nil {
			log.Fatalf("Failed to start enrichment service: %v", err)
			return
		}*/

	// Set up syslog channel and handler
	channel := make(syslog.LogPartsChannel)
	chanHandler := syslog.NewChannelHandler(channel)

	// Initialize message handler with Neo4j client
	messageHandler := handler.New(ctx, neo4jClient, enricher, channel)

	// Configure and start syslog server
	server := syslog.NewServer()
	server.SetFormat(syslog.Automatic)
	server.SetHandler(chanHandler)

	if err := server.ListenUDP("0.0.0.0:514"); err != nil {
		log.Fatalf("Failed to listen on UDP port 514: %v", err)
		return
	}
	if err := server.Boot(); err != nil {
		log.Fatalf("Failed to start syslog server: %v", err)
		return
	}

	// Start message handler
	if err := messageHandler.Start(); err != nil {
		log.Fatalf("Failed to start message handler: %v", err)
		return
	}

	server.Wait()
}

// setupNeo4jConnection initializes the Neo4j client with configuration from environment variables
func setupNeo4jConnection(ctx context.Context) (*database.Neo4jClient, error) {
	// Get Neo4j connection parameters from environment variables
	uri := getEnvWithDefault("NEO4J_URI", "neo4j://192.168.0.223:7687")
	username := getEnvWithDefault("NEO4J_USERNAME", "neo4j")
	password := getEnvWithDefault("NEO4J_PASSWORD", "123412341234")
	db := getEnvWithDefault("NEO4J_DATABASE", "neo4j")

	log.Printf("Connecting to Neo4j at %s", uri)

	// Create Neo4j configuration
	config := database.Neo4jConfig{
		URI:      uri,
		Username: username,
		Password: password,
		Database: db,
	}

	// Initialize and return Neo4j client
	return database.NewNeo4jClient(ctx, config)
}

// getEnvWithDefault returns the value of an environment variable or a default value if not set
func getEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
