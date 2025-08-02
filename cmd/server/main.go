package main

import (
	"context"
	"errors"
	"log"
	"log/slog"

	"github.com/EduardoOliveira/ckc/enrichment"
	"github.com/EduardoOliveira/ckc/handler"
	"github.com/EduardoOliveira/ckc/internal/cfg"
	"github.com/EduardoOliveira/ckc/internal/ptr"
	"github.com/EduardoOliveira/ckc/neo4j"
	"github.com/EduardoOliveira/ckc/types"
	syslog "gopkg.in/mcuadros/go-syslog.v2"
)

func main() {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)

	// Initialize Neo4j connection
	nClient := mustSetupNeo4jClient(ctx)
	defer nClient.Close(ctx)
	slog.Info("Connected to Neo4j", "uri", cfg.Must("NEO4J_URI"), "database", cfg.Must("NEO4J_DATABASE"))

	handler := handler.New(ctx,
		map[types.ServiceName][]handler.ContentParser{
			types.SSHDService: {
				ptr.To(handler.NewSSHDParser()),
			},
		},
		map[types.ServiceName][]handler.ContentStore{
			types.SSHDService: {
				ptr.To(neo4j.NewNeo4jSSHD(nClient)),
			},
		},
		map[types.ServiceName][]handler.ContentEnricher{
			types.SSHDService: {
				ptr.To(enrichment.NewAIPDBEnricher(ctx, cfg.Must("AIPDB_API_KEY"), nClient)),
			},
		},
	)

	mustRunRsyslogServer(cancel, handler)

	select {
	case <-ctx.Done():
		log.Println("Shutting down gracefully...")
	}
	cancel(nil)
}

func mustSetupNeo4jClient(ctx context.Context) *neo4j.Neo4jClient {
	config := neo4j.Neo4jConfig{
		URI:      cfg.Must("NEO4J_URI"),
		Username: cfg.Must("NEO4J_USERNAME"),
		Password: cfg.Must("NEO4J_PASSWORD"),
		Database: cfg.Must("NEO4J_DATABASE"),
	}
	client, err := neo4j.NewNeo4jClient(ctx, config)
	if err != nil {
		panic("Failed to create Neo4j client: " + err.Error())
	}
	return client
}

func mustRunRsyslogServer(cancel context.CancelCauseFunc, handler *handler.Handler) syslog.LogPartsChannel {
	channel := make(syslog.LogPartsChannel)
	server := syslog.NewServer()
	server.SetFormat(syslog.Automatic)
	server.SetHandler(handler)

	if err := server.ListenUDP(cfg.Must("RSYSLOG_SERVER")); err != nil {
		panic("Failed to start syslog server: " + err.Error())
	}
	if err := server.Boot(); err != nil {
		panic("Failed to start syslog server: " + err.Error())
	}
	go func() {
		server.Wait()
		cancel(errors.New("ryslog server stoped")) // Notify the main function to shut down
	}()
	return channel
}
