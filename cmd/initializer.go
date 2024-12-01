package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/merbinr/deduplicator/internal/config"
)

func initializer() {
	err := config.LoadConfig()
	if err != nil {
		slog.Error(fmt.Sprintf("unable to load config file, %s", err))
		os.Exit(1)
	}
}
