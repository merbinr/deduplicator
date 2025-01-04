package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/merbinr/deduplicator/internal/config"
	opensearchhelper "github.com/merbinr/deduplicator/internal/opensearch_helper"
	"github.com/merbinr/deduplicator/internal/outputChannel"
	"github.com/merbinr/deduplicator/internal/queue/incoming"
	"github.com/merbinr/deduplicator/internal/queue/outgoing"
	"github.com/merbinr/deduplicator/pkg/logger"
)

func main() {
	// Getting logger instance
	logger := logger.GetLogger()

	logger.Info("Initializing deduplication process")
	initializer()

	logger.Info("De-duplication process started")
	go incoming.ConsumeMessage()

	if config.Config.OutputMethod == "queue" {
		for msg := range outputChannel.OutputChannel {
			err := outgoing.SendMessage(&msg)
			if err != nil {
				logger.Error(fmt.Sprintf("unable to send message to outgoing queue, err: %s", err))
			}
		}
	} else if config.Config.OutputMethod == "opensearch" {
		var wg sync.WaitGroup
		wg.Add(1)
		go opensearchhelper.IngestLogs()
		wg.Wait()

	} else {
		logger.Error(fmt.Sprintf("Invalid output method: %s", config.Config.OutputMethod))
		os.Exit(1)
	}
}
