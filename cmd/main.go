package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/merbinr/deduplicator/internal/config"
	"github.com/merbinr/deduplicator/internal/opensearch_helper"
	outputchannel "github.com/merbinr/deduplicator/internal/output_channel"
	"github.com/merbinr/deduplicator/internal/queue/incoming"
	"github.com/merbinr/deduplicator/internal/queue/outgoing"
)

func main() {
	slog.Info("Initializing deduplication process")
	initializer()

	fmt.Println("De-duplication process started")
	go incoming.ConsumeMessage()

	if config.Config.OutputMethod == "queue" {
		for msg := range outputchannel.OutputChannel {
			err := outgoing.SendMessage(msg)
			if err != nil {
				slog.Error(fmt.Sprintf("unable to send message to outgoing queue, err: %s", err))
			}
		}
	} else if config.Config.OutputMethod == "webhook" {
		for {
			go opensearch_helper.IngestLogs()
		}
	} else {
		slog.Error(fmt.Sprintf("Invalid output method: %s", config.Config.OutputMethod))
		os.Exit(1)
	}
}
