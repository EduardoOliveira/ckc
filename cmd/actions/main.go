package main

import (
	"context"
	"flag"

	"github.com/EduardoOliveira/ckc/enrichment"
	"github.com/EduardoOliveira/ckc/internal/cfg"
	"github.com/EduardoOliveira/ckc/internal/ptr"
	"github.com/EduardoOliveira/ckc/neo4j"
)

func main() {
	ip := ptr.Val(flag.Bool("eips", true, "Enrich IPs"))
	_ = ptr.Val(flag.Bool("eusernames", true, "Enrich usernames"))

	flag.Parse()

	if len(flag.Args()) == 0 {
		flag.Usage()
		return
	}

	nClient := neo4j.MustSetupNeo4jClient(context.Background())

	if ip {
		aipdbEnricher := enrichment.NewAIPDBEnricher(context.Background(), cfg.Must("AIPDB_API_KEY"), nClient)
		if err := aipdbEnricher.EnrichAll(context.Background()); err != nil {
			panic("Failed to enrich IPs: " + err.Error())
		}

	}
}
