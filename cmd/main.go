package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/caarlos0/env/v11"

	"reco/pkg/client"
	"reco/pkg/poller"
	"reco/pkg/storage"
)

type config struct {
	AsanaBaseUrl   string `env:"ASANA_BASE_URL,required"`
	AsanaToken     string `env:"ASANA_TOKEN,required"`
	AsanaWorkspace string `env:"ASANA_WORKSPACE,required"`
	Frequency      int    `env:"FREQUENCY,required"`
	Attempts       int    `env:"ATTEMPTS" envDefault:"5"`
	PageLimit      int    `env:"PAGE_LIMIT" envDefault:"100"`
}

func main() {
	projectsStorage := storage.New("projects.jsonl")
	usersStorage := storage.New("users.jsonl")

	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		slog.Error("config parse failed", "error", err)
		os.Exit(1)
	}

	if cfg.Frequency < 1 {
		slog.Error("invalid config", "field", "FREQUENCY", "error", "must be greater than 0")
		os.Exit(1)
	}

	if cfg.Attempts < 1 {
		slog.Error("invalid config", "field", "ATTEMPTS", "error", "must be greater than 0")
		os.Exit(1)
	}

	if cfg.PageLimit < 1 || cfg.PageLimit > 100 {
		slog.Error("invalid config", "field", "PAGE_LIMIT", "error", "must be between 1 and 100")
		os.Exit(1)
	}

	slog.Info(
		"extractor started",
		"frequency_seconds", cfg.Frequency,
		"attempts", cfg.Attempts,
		"page_limit", cfg.PageLimit,
	)

	asanaClient := client.New(cfg.AsanaBaseUrl, cfg.AsanaToken)
	every := time.Duration(cfg.Frequency) * time.Second
	asanaPoller := poller.New(
		asanaClient,
		cfg.AsanaWorkspace,
		projectsStorage,
		usersStorage,
		every,
		cfg.Attempts,
		cfg.PageLimit,
	)

	slog.Info("polling projects")
	if err := asanaPoller.PollProjects(); err != nil {
		slog.Error("projects poll failed", "error", err)
		os.Exit(1)
	}

	slog.Info("projects appended")

	slog.Info("polling users")
	if err := asanaPoller.PollUsers(); err != nil {
		slog.Error("users poll failed", "error", err)
		os.Exit(1)
	}

	slog.Info("users appended")
}
