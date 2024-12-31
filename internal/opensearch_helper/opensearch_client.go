package opensearch_helper

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/merbinr/deduplicator/internal/config"
	outputchannel "github.com/merbinr/deduplicator/internal/output_channel"
	"github.com/opensearch-project/opensearch-go"
)

var OpenSearchClient *opensearch.Client

func LoadOpenSearchClient() {
	username := config.Config.Opensearch.Username
	port := config.Config.Opensearch.Port
	host := config.Config.Opensearch.Host
	password := os.Getenv("DEDUPLICATOR_OPENSEARCH_PASSWORD")
	if password == "" {
		slog.Error("DEDUPLICATOR_OPENSEARCH_PASSWORD env is empty, please set it")
		os.Exit(1)
	}

	connString := fmt.Sprintf("https://%s:%d/", host, port)
	client, err := opensearch.NewClient(opensearch.Config{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Addresses: []string{connString},
		Username:  username, // For testing only. Don't store credentials in code.
		Password:  password,
	})
	if err != nil {
		log.Fatalf("Error creating the client: %s", err)
	}
	OpenSearchClient = client
	createIndexes()
}

func createIndexes() {
	createAwsVpcIndexIfNotExists()
}

func ingestIntoOpensearch(logs *[]byte) {
	blk, err := OpenSearchClient.Bulk(bytes.NewReader(*logs))
	if err != nil {
		log.Fatalf("Error ingesting data: %s", err)
	}
	if blk.IsError() {
		log.Fatalf("Error ingesting data, status code: %d, body: %s", blk.StatusCode,
			blk.String())
	}
	slog.Info("Data ingested successfully")
}

func IngestLogs() {
	currentRetries := 0

	if len(outputchannel.OutputChannel) == 0 {
		time.Sleep(500 * time.Millisecond)
		slog.Debug("No logs to ingest")
		return

	} else if len(outputchannel.OutputChannel) >= config.Config.Opensearch.PreferedBatchSize {
		slog.Debug(fmt.Sprintf("More logs to ingest than prefered batch size: %d", len(outputchannel.OutputChannel)))
		preparedLogs := prepareAwsVpcLogsForIngestion(config.Config.Opensearch.PreferedBatchSize)
		ingestIntoOpensearch(&preparedLogs)

	} else {
		slog.Info("Batch size is less than prefered batch size")
		if currentRetries < config.Config.Opensearch.Retries {
			slog.Debug("As batch size is less than prefered batch size, waiting for more logs")
			time.Sleep(time.Duration(config.Config.Opensearch.RetryDelay) * time.Second)
			currentRetries++

		} else {
			slog.Info("Waiting exceeded, ingesting logs even if batch size is less than prefered batch size")
			preparedLogs := prepareAwsVpcLogsForIngestion(len(outputchannel.OutputChannel))
			ingestIntoOpensearch(&preparedLogs)
		}
	}
}
