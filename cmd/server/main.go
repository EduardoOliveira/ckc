package main

import (
	"context"
	"errors"
	"log"
	"log/slog"

	"github.com/EduardoOliveira/ckc/handler"
	"github.com/EduardoOliveira/ckc/internal/cfg"
	"github.com/EduardoOliveira/ckc/internal/ptr"
	"github.com/EduardoOliveira/ckc/stores"
	"github.com/EduardoOliveira/ckc/types"
	syslog "gopkg.in/mcuadros/go-syslog.v2"
)

func main() {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)

	// Initialize Neo4j connection
	neo4j, err := mustSetupNeo4jClient(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to Neo4j: %v", err)
	}
	defer neo4j.Close(ctx)
	slog.Info("Connected to Neo4j", "uri", cfg.Must("NEO4J_URI"), "database", cfg.Must("NEO4J_DATABASE"))

	handler := handler.New(ctx,
		map[types.ServiceName][]handler.ContentParser{
			types.SSHDService: {
				ptr.To(handler.NewSSHDParser()),
			},
		},
		map[types.ServiceName][]handler.ContentStore{
			types.SSHDService: {
				ptr.To(stores.NewNeo4jSSHD(neo4j)),
			},
		},
		map[types.ServiceName][]handler.ContentEnricher{},
	)

	mustRunRsyslogServer(cancel, handler)

	select {
	case <-ctx.Done():
		log.Println("Shutting down gracefully...")
	}
	cancel(nil)
}

func mustSetupNeo4jClient(ctx context.Context) (*stores.Neo4jClient, error) {
	config := stores.Neo4jConfig{
		URI:      cfg.Must("NEO4J_URI"),
		Username: cfg.Must("NEO4J_USERNAME"),
		Password: cfg.Must("NEO4J_PASSWORD"),
		Database: cfg.Must("NEO4J_DATABASE"),
	}

	return stores.NewNeo4jClient(ctx, config)
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
