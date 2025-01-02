package opensearchhelper

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/merbinr/deduplicator/internal/config"
	"github.com/merbinr/deduplicator/internal/outputChannel"
	"github.com/opensearch-project/opensearch-go"
)

var OpenSearchClient *opensearch.Client

func LoadOpenSearchClient() {
	username := config.Config.Services.Opensearch.Username
	port := config.Config.Services.Opensearch.Port

	DEDUPLICATOR_OPENSEARCH_HOST_ENV := fmt.Sprintf("%s_DEDUPLICATOR_OPENSEARCH_HOST",
		config.Config.StageName)

	host := os.Getenv(DEDUPLICATOR_OPENSEARCH_HOST_ENV)
	if host == "" {
		slog.Error(fmt.Sprintf("%s env is empty, please set it", DEDUPLICATOR_OPENSEARCH_HOST_ENV))
		os.Exit(1)
	}

	DEDUPLICATOR_OPENSEARCH_PASSWORD_ENV := fmt.Sprintf("%s_DEDUPLICATOR_OPENSEARCH_PASSWORD",
		config.Config.StageName)
	password := os.Getenv(DEDUPLICATOR_OPENSEARCH_PASSWORD_ENV)
	if password == "" {
		slog.Error(fmt.Sprintf("%s env is empty, please set it", DEDUPLICATOR_OPENSEARCH_PASSWORD_ENV))
		os.Exit(1)
	}

	connString := fmt.Sprintf("http://%s:%d/", host, port)
	client, err := opensearch.NewClient(opensearch.Config{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Addresses: []string{connString},
		Username:  username, // For testing only. Don't store credentials in code.
		Password:  password,
	})
	if err != nil {
		slog.Error(fmt.Sprintf("Error creating the client: %s", err))
		os.Exit(1)
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
		slog.Error(fmt.Sprintf("Error ingesting data: %s", err))
		os.Exit(1)
	}
	if blk.IsError() {
		slog.Error(fmt.Sprintf("Error ingesting data, status code: %d, body: %s", blk.StatusCode,
			blk.String()))
		os.Exit(1)
	}
	slog.Info("Data ingested successfully")
}

func IngestLogs() {
	currentRetries := 0

	if len(outputChannel.OutputChannel) == 0 {
		time.Sleep(500 * time.Millisecond)
		slog.Debug("No logs to ingest")
		return

	} else if len(outputChannel.OutputChannel) >= config.Config.Services.Opensearch.PreferredBatchSize {
		slog.Debug(fmt.Sprintf("More logs to ingest than prefered batch size: %d", len(outputChannel.OutputChannel)))
		preparedLogs := prepareAwsVpcLogsForIngestion(config.Config.Services.Opensearch.PreferredBatchSize)
		ingestIntoOpensearch(&preparedLogs)

	} else {
		slog.Info("Batch size is less than prefered batch size")
		if currentRetries < config.Config.Services.Opensearch.Retries {
			slog.Debug("As batch size is less than prefered batch size, waiting for more logs")
			time.Sleep(time.Duration(config.Config.Services.Opensearch.RetryDelay) * time.Second)
			currentRetries++

		} else {
			slog.Info("Waiting exceeded, ingesting logs even if batch size is less than prefered batch size")
			preparedLogs := prepareAwsVpcLogsForIngestion(len(outputChannel.OutputChannel))
			ingestIntoOpensearch(&preparedLogs)
		}
	}
}
