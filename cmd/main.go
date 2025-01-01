package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/merbinr/deduplicator/internal/config"
	opensearchhelper "github.com/merbinr/deduplicator/internal/opensearch_helper"
	"github.com/merbinr/deduplicator/internal/outputChannel"
	"github.com/merbinr/deduplicator/internal/queue/incoming"
	"github.com/merbinr/deduplicator/internal/queue/outgoing"
)

func main() {
	slog.Info("Initializing deduplication process")
	initializer()

	fmt.Println("De-duplication process started")
	go incoming.ConsumeMessage()

	if config.Config.OutputMethod == "queue" {
		for msg := range outputChannel.OutputChannel {
			err := outgoing.SendMessage(msg)
			if err != nil {
				slog.Error(fmt.Sprintf("unable to send message to outgoing queue, err: %s", err))
			}
		}
	} else if config.Config.OutputMethod == "webhook" {
		for {
			go opensearchhelper.IngestLogs()
		}
	} else {
		slog.Error(fmt.Sprintf("Invalid output method: %s", config.Config.OutputMethod))
		os.Exit(1)
	}
}
