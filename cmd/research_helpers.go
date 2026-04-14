package cmd

import (
	"context"

	"github.com/8op-org/gl1tch/internal/esearch"
	"github.com/8op-org/gl1tch/internal/provider"
	"github.com/8op-org/gl1tch/internal/research"
)

func buildToolLoop(repoPath string) (*research.ToolLoop, error) {
	cfg, _ := loadConfig()

	// ES client (optional — graceful if not available)
	var es *esearch.Client
	esClient := esearch.NewClient("http://localhost:9200")
	if err := esClient.Ping(context.Background()); err == nil {
		es = esClient
	}

	// Telemetry (nil-safe if no ES)
	tel := esearch.NewTelemetry(es)
	if tel != nil {
		tel.EnsureIndices(context.Background())
	}

	// Tools
	tools := research.NewToolSet(repoPath, es)

	// Tiered runner
	tiers := cfg.Tiers
	if len(tiers) == 0 {
		tiers = provider.DefaultTiers()
	}
	runner := provider.NewTieredRunner(tiers, providerReg)
	runner.Resolver = cfg.BuildProviderResolver()

	return research.NewToolLoop(tools, runner, tel), nil
}
