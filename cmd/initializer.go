package main

import (
	"fmt"
	"os"

	"github.com/merbinr/deduplicator/internal/config"
	opensearchhelper "github.com/merbinr/deduplicator/internal/opensearch_helper"
	"github.com/merbinr/deduplicator/internal/queue/incoming"
	"github.com/merbinr/deduplicator/internal/queue/outgoing"
	rediscache "github.com/merbinr/deduplicator/internal/redis-cache"
	"github.com/merbinr/deduplicator/pkg/logger"
)

func initializer() {
	// Getting logger instance
	logger := logger.GetLogger()

	// Loading config
	err := config.LoadConfig()
	if err != nil {
		logger.Error(fmt.Sprintf("unable to load config file, %s", err))
		os.Exit(1)
	}

	// Loading incomming queue
	err = incoming.CreateQueueClient()
	if err != nil {
		logger.Error(fmt.Sprintf("unable to create incomming queue client, %s", err))
		os.Exit(1)
	}

	if config.Config.OutputMethod == "queue" {
		// Loading outgoing queue
		err = outgoing.CreateQueueClient()
		if err != nil {
			logger.Error(fmt.Sprintf("unable to create outgoing queue client, %s", err))
			os.Exit(1)
		}
	} else if config.Config.OutputMethod == "opensearch" {
		// Load opensearch client
		opensearchhelper.LoadOpenSearchClient()
	} else {
		logger.Error(fmt.Sprintf("Invalid output method: %s", config.Config.OutputMethod))
		os.Exit(1)
	}

	// Load redis cache
	err = rediscache.LoadRedisClient()
	if err != nil {
		logger.Error(fmt.Sprintf("unable to redis cache client, %s", err))
		os.Exit(1)
	}
}
