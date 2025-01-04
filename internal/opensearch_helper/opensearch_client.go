package opensearchhelper

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/merbinr/deduplicator/internal/config"
	"github.com/merbinr/deduplicator/internal/outputChannel"
	"github.com/merbinr/deduplicator/pkg/logger"
	"github.com/opensearch-project/opensearch-go"
)

var OpenSearchClient *opensearch.Client

func LoadOpenSearchClient() {
	logger := logger.GetLogger()
	username := config.Config.Services.Opensearch.Username
	port := config.Config.Services.Opensearch.Port

	DEDUPLICATOR_OPENSEARCH_HOST_ENV := fmt.Sprintf("%s_DEDUPLICATOR_OPENSEARCH_HOST",
		config.Config.StageName)

	host := os.Getenv(DEDUPLICATOR_OPENSEARCH_HOST_ENV)
	if host == "" {
		logger.Error(fmt.Sprintf("%s env is empty, please set it", DEDUPLICATOR_OPENSEARCH_HOST_ENV))
		os.Exit(1)
	}

	DEDUPLICATOR_OPENSEARCH_PASSWORD_ENV := fmt.Sprintf("%s_DEDUPLICATOR_OPENSEARCH_PASSWORD",
		config.Config.StageName)
	password := os.Getenv(DEDUPLICATOR_OPENSEARCH_PASSWORD_ENV)
	if password == "" {
		logger.Error(fmt.Sprintf("%s env is empty, please set it", DEDUPLICATOR_OPENSEARCH_PASSWORD_ENV))
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
		logger.Error(fmt.Sprintf("Error creating the client: %s", err))
		os.Exit(1)
	}
	OpenSearchClient = client
	createIndexes()
}

func createIndexes() {
	createAwsVpcIndexIfNotExists()
}

func ingestIntoOpensearch(logs *[]byte) {
	logger := logger.GetLogger()
	blk, err := OpenSearchClient.Bulk(bytes.NewReader(*logs))
	if err != nil {
		logger.Error(fmt.Sprintf("Error ingesting data: %s", err))
		os.Exit(1)
	}
	if blk.IsError() {
		logger.Error(fmt.Sprintf("Error ingesting data, status code: %d, body: %s", blk.StatusCode,
			blk.String()))
		os.Exit(1)
	}
	logger.Info("Data ingested successfully")
}

func IngestLogs() {
	logger := logger.GetLogger()
	currentRetries := 0
	for {
		logger.Info(fmt.Sprintf("Trying logs ingest into OpenSearch, current batch size: %d",
			len(outputChannel.OutputChannel)))
		if len(outputChannel.OutputChannel) == 0 {
			logger.Debug("No logs to ingest, waiting 5 seconds")
			time.Sleep(5 * time.Second)

		} else if len(outputChannel.OutputChannel) >= config.Config.Services.Opensearch.PreferredBatchSize {
			logger.Debug(fmt.Sprintf("More logs to ingest than prefered batch size: %d, preffered batch size: %d",
				len(outputChannel.OutputChannel), config.Config.Services.Opensearch.PreferredBatchSize))
			preparedLogs := prepareAwsVpcLogsForIngestion(&config.Config.Services.Opensearch.PreferredBatchSize)
			ingestIntoOpensearch(&preparedLogs)
			currentRetries = 0
			time.Sleep(200 * time.Millisecond)

		} else {
			logger.Info("Batch size is less than prefered batch size")
			if currentRetries < config.Config.Services.Opensearch.Retries {
				logger.Debug(fmt.Sprintf("As batch size is less than prefered batch size, waiting for more logs, remaining retries: %d",
					config.Config.Services.Opensearch.Retries-currentRetries))
				time.Sleep(time.Duration(config.Config.Services.Opensearch.RetryDelay) * time.Millisecond)
				currentRetries++

			} else {
				logger.Info("Waiting exceeded, ingesting logs even if batch size is less than prefered batch size")
				batchSize := len(outputChannel.OutputChannel)
				preparedLogs := prepareAwsVpcLogsForIngestion(&batchSize)
				ingestIntoOpensearch(&preparedLogs)
				currentRetries = 0
			}
		}
	}
}
