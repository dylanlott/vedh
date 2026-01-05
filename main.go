package main

import (
	"log/slog"
	"os"
	"strings"

	"github.com/zeebo/errs"

	"github.com/openmtg/edh-go/persistence"
	"github.com/openmtg/edh-go/server"

	"github.com/kelseyhightower/envconfig"
)

func main() {
	var level slog.Level
	level = slog.LevelInfo
	if envLevel := strings.TrimSpace(os.Getenv("LOG_LEVEL")); envLevel != "" {
		if err := level.UnmarshalText([]byte(strings.ToLower(envLevel))); err != nil {
			// default to INFO if invalid
			level = slog.LevelInfo
		}
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
	slog.SetDefault(logger)

	var cfg server.Conf
	err := envconfig.Process("", &cfg)
	if err != nil {
		logger.Error("failed to load config", "err", err)
		os.Exit(1)
	}
	db, err := persistence.NewDB(cfg.PostgresURL)
	if err != nil {
		logger.Error("failed to connect database", "err", errs.Wrap(err))
		os.Exit(1)
	}
	logger.Info("successfully opened database connection")
	s, err := server.NewGraphQLServer(db, cfg, logger)
	if err != nil {
		logger.Error("failed to create graphql server", "err", err)
		os.Exit(1)
	}
	logger.Info("graphQL server attempting to start")
	err = s.Serve("/graphql", cfg.DefaultPort)
	if err != nil {
		logger.Error("server exited", "err", err)
		os.Exit(1)
	}
	logger.Info("serving", "graphql", "/graphql", "port", cfg.DefaultPort)
	logger.Info("serving", "playground", "/playground", "port", cfg.DefaultPort)
}
