package opensearchhelper

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"strings"

	"github.com/merbinr/deduplicator/internal/outputChannel"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
)

func createAwsVpcIndexIfNotExists() {
	alreadyExists := checkAwsVpcIndexExists()
	if !alreadyExists {
		createAwsVpcIndex()
	}
}

func createAwsVpcIndex() {
	mapping := strings.NewReader(`
	{
	  "mappings": {
	    "properties": {
	      "Cloud": {
	        "type": "keyword"
	      },
	      "Type": {
	        "type": "keyword"
	      },
	      "Version": {
	        "type": "integer"
	      },
	      "AccountID": {
	        "type": "keyword"
	      },
	      "InterfaceID": {
	        "type": "keyword"
	      },
	      "SourceIP": {
	        "type": "keyword"
	      },
	      "DestinationIP": {
	        "type": "keyword"
	      },
	      "DestinationPort": {
	        "type": "integer"
	      },
	      "SourcePort": {
	        "type": "integer"
	      },
	      "Protocol": {
	        "type": "integer"
	      },
	      "Packets": {
	        "type": "integer"
	      },
	      "Bytes": {
	        "type": "integer"
	      },
	      "StartTime": {
	        "type": "date",
	        "format": "epoch_second"
	      },
	      "EndTime": {
	        "type": "date",
	        "format": "epoch_second"
	      },
	      "Action": {
	        "type": "keyword"
	      },
	      "LogStatus": {
	        "type": "keyword"
	      }
	    }
	  }
	}`)

	req := opensearchapi.IndicesCreateRequest{
		Index: "aws_vpc",
		Body:  mapping,
	}
	response, err := req.Do(context.Background(), OpenSearchClient)
	if err != nil {
		log.Fatalf("Error creating index: %s", err)
	}

	if response.StatusCode != 200 {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			log.Fatalf("Error reading response body: %v\n", err)
		}
		log.Fatalf("Error creating index, status code: %d, body: %s",
			response.StatusCode, string(body))
	} else {
		log.Println("Index created successfully")
	}
}

func checkAwsVpcIndexExists() bool {
	indexName := "aws_vpc"
	res, err := OpenSearchClient.Indices.Exists([]string{indexName})
	if err != nil {
		log.Fatalf("Error checking if index exists: %s", err)
	}
	if res.StatusCode == 200 {
		log.Println("Index exists, skipping creation")
		return true
	} else if res.StatusCode == 404 {
		log.Println("Index doesn't exist")
		return false
	} else {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			log.Fatalf("Error reading response body: %v\n", err)
		}
		log.Fatalf("Error checking if index exists, status code: %d, body %s",
			res.StatusCode, string(body))
		return false
	}
}

func prepareAwsVpcLogsForIngestion(batchSize int) []byte {
	var buf bytes.Buffer
	indexName := "aws_vpc"
	currentRecord := 1

	for log := range outputChannel.OutputChannel {
		meta := map[string]interface{}{
			"index": map[string]string{"_index": indexName},
		}
		metaJSON, _ := json.Marshal(meta)
		buf.Write(metaJSON)
		buf.WriteByte('\n')

		buf.Write(log)
		buf.WriteByte('\n')

		// Getting messages from channel is forever running loop, So we need to break it after batchSize
		if currentRecord == batchSize {
			break
		}
		currentRecord++
	}
	bytes := buf.Bytes()
	return bytes
}
