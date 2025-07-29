package handler

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/EduardoOliveira/ckc/types"
	syslogformat "gopkg.in/mcuadros/go-syslog.v2/format"
)

type ContentParser interface {
	Name() string
	Parse(ctx context.Context, content string, parent types.ParsedEvent) (types.ParsedEvent, error)
}

type ContentEnricher interface {
	Name() string
	Enrich(ctx context.Context, parsed types.ParsedEvent)
}

type ContentStore interface {
	Name() string
	Store(ctx context.Context, parsed types.ParsedEvent) error
}

type Handler struct {
	ctx       context.Context
	stores    map[types.ServiceName][]ContentStore
	parsers   map[types.ServiceName][]ContentParser
	enrichers map[types.ServiceName][]ContentEnricher
	now       func() time.Time // Function to get the current time, can be overridden for testing
}

func New(ctx context.Context,
	parsers map[types.ServiceName][]ContentParser,
	stores map[types.ServiceName][]ContentStore,
	enrichers map[types.ServiceName][]ContentEnricher,
) *Handler {
	return &Handler{
		ctx:       ctx,
		parsers:   parsers,
		stores:    stores,
		enrichers: enrichers,
		now:       time.Now,
	}
}

func (h *Handler) Handle(logParts syslogformat.LogParts, _ int64, err error) {
	if err != nil {
		slog.Error("Error handling log parts: ", "error", err)
		return
	}

	var ok bool
	var content string
	if content = getStringValue(logParts, "content", ""); content == "" {
		slog.Warn("Log parts missing 'content', skipping", "logParts", logParts)
		return
	}

	var serviceName types.ServiceName
	if serviceName, ok = types.ParseServiceNameFromAny(logParts["tag"]); !ok {
		slog.Warn("Failed to parse service name from log parts", "logParts", logParts)
		return
	}

	slog.Info("Handling log parts", "service", serviceName, slog.Any("logParts", logParts))
	parsed := types.ParsedEvent{
		Hostname:    getStringValue(logParts, "hostname", ""),
		Ingestion:   getTimeFromLogParts(logParts),
		ServiceName: serviceName,
	}

	ctx, cancel := context.WithTimeout(h.ctx, 5*time.Second)
	defer cancel()

	if len(h.parsers[serviceName]) == 0 {
		slog.Warn("No parsers registered for service", "service", serviceName)
	} else {
		for _, parser := range h.parsers[serviceName] {
			if parser == nil {
				slog.Warn("No parser found for service", "service", serviceName)
				return
			}
			parsed, err = parser.Parse(ctx, content, parsed)
			if err != nil {
				slog.Warn("Failed to parse content for service", "service", serviceName, "parser", parser.Name(), "error", err)
				return
			}
		}
	}

	if len(h.stores[serviceName]) == 0 {
		slog.Warn("No stores registered for service", "service", serviceName)
		return
	} else {
		for _, store := range h.stores[serviceName] {
			if store == nil {
				slog.Warn("No store found for service", "service", serviceName)
				return
			}
			if err := store.Store(ctx, parsed); err != nil {
				slog.Error("Failed to store parsed event", "service", serviceName, "store", store.Name(), "error", err)
				return
			}
		}
	}

	if len(h.enrichers[serviceName]) == 0 {
		slog.Warn("No enrichers registered for service", "service", serviceName)
		return
	} else {
		for _, enricher := range h.enrichers[serviceName] {
			if enricher == nil {
				slog.Warn("No enricher found for service", "service", serviceName)
				return
			}
			// enrichment should be done asynchronously
			go enricher.Enrich(ctx, parsed)
		}
	}
}

// getTimestampFromLogParts extracts timestamp from log parts
func getTimeFromLogParts(logParts map[string]any) time.Time {
	// Try to get timestamp from log parts
	if ts, ok := logParts["timestamp"]; ok {
		switch v := ts.(type) {
		case string:
			// Layout      = "01/02 03:04:05PM '06 -0700" // The reference time, in numerical order.
			t, err := time.Parse("2006-01-02 15:04:05 -0700 MST", v)
			if err == nil {
				return t
			}
			slog.Warn("Failed to parse timestamp from string", "timestamp", v, "error", err)
		case time.Time:
			return v
		case int64:
			return time.Unix(v, 0)
		case int:
			return time.Unix(int64(v), 0)
		}
	}

	// If no timestamp found or not a valid type, use current time
	return time.Now()
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
