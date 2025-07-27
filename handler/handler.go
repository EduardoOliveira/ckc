package handler

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/EduardoOliveira/ckc/database"
	"github.com/EduardoOliveira/ckc/enrichment"
	"github.com/EduardoOliveira/ckc/types"
	syslog "gopkg.in/mcuadros/go-syslog.v2"
	syslogformat "gopkg.in/mcuadros/go-syslog.v2/format"
)

type Handler struct {
	channel     syslog.LogPartsChannel
	ctx         context.Context
	neo4jClient *database.Neo4jClient
	enricher    *enrichment.Enrichment
	sshdHandler sshdHandler
	now         func() time.Time // Function to get the current time, can be overridden for testing
}

func New(ctx context.Context, neo4jClient *database.Neo4jClient, enrichment *enrichment.Enrichment, channel syslog.LogPartsChannel) *Handler {
	return &Handler{
		channel:     channel,
		ctx:         ctx,
		neo4jClient: neo4jClient,
		enricher:    enrichment,
		sshdHandler: NewSSHDHandler(),
		now:         time.Now,
	}
}

func (h *Handler) Start() error {
	go func() {
		for {
			select {
			case logParts := <-h.channel:
				go h.handleMessage(h.ctx, logParts)
			case <-h.ctx.Done():
				return
			}
		}
	}()

	return nil
}

func (h *Handler) handleMessage(_ context.Context, logParts syslogformat.LogParts) {
	// fmt.Println(logParts)
	fmt.Println(getStringValue(logParts, "content", ""))
	content := getStringValue(logParts, "content", "")
	var cyphers []types.Cypher
	if content == "" {
		return
	}
	switch getStringValue(logParts, "tag", "") {
	case "sshd":
		// Handle SSH logs
		_, c, err := h.sshdHandler.Parse(content)
		if err != nil {
			log.Printf("Failed to parse SSH log: %v", err)
			return
		}
		if len(c) > 0 {
			cyphers = append(cyphers, c...)
		}
		// h.enricher.EnrichIP(ip)
	}
	if len(cyphers) == 0 {
		log.Printf("No Cypher objects created for log: %v", logParts)
		return
	}
	out, err := h.neo4jClient.Write(h.ctx, cyphers...)
	if err != nil {
		log.Printf("Failed to write to Neo4j: %v", err)
		return
	}
	log.Printf("Write result: %v", out)
}

// getTimestampFromLogParts extracts timestamp from log parts
func getTimestampFromLogParts(logParts map[string]any) int64 {
	// Try to get timestamp from log parts
	if ts, ok := logParts["timestamp"]; ok {
		switch t := ts.(type) {
		case time.Time:
			return t.Unix()
		case int64:
			return t
		case int:
			return int64(t)
		}
	}

	// If no timestamp found or not a valid type, use current time
	return time.Now().Unix()
}

// getStringValue safely extracts a string value from map
func getStringValue(data map[string]any, key, defaultValue string) string {
	if value, ok := data[key]; ok {
		if strValue, ok := value.(string); ok {
			return strValue
		}
		// Try to convert to string
		return fmt.Sprintf("%v", value)
	}
	return defaultValue
}
