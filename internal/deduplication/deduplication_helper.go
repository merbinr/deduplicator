package deduplication

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"reflect"
	"strings"

	"github.com/buger/jsonparser"
	"github.com/merbinr/deduplicator/internal/config"
	outputchannel "github.com/merbinr/deduplicator/internal/output_channel"
	rediscache "github.com/merbinr/deduplicator/internal/redis_cache"
	"github.com/merbinr/log_models/models"
)

func ProcessDeduplication(msg []byte) error {
	cloud, err := jsonparser.GetString(msg, "Cloud")
	slog.Debug(fmt.Sprintf("Cloud: %s", cloud))
	if err != nil {
		return fmt.Errorf("unable to get cloud value from log message, err: %s", err)
	}

	log_type, err := jsonparser.GetString(msg, "Type")
	slog.Debug(fmt.Sprintf("Log Type : %s", log_type))
	if err != nil {
		return fmt.Errorf("unable to get log_type value from log message")
	}

	if cloud == "aws" && log_type == "vpc" {
		slog.Debug("As cloud is aws and log type is vpc, processing AWS VPC log")
		err = processAwsVpcLogs(msg)
		if err != nil {
			return fmt.Errorf("unable to process AWS VPC log, err: %s", err)
		}
	}
	if err != nil {
		return fmt.Errorf("unable to acknowlege the message, err: %s", err)
	}
	return nil
}

func processAwsVpcLogs(vpc_log_msg []byte) error {
	var vpc_log_data models.VpcNormalizedData
	slog.Debug("Unmarshalling the log message to struct VpcNormalizedData")
	err := json.Unmarshal(vpc_log_msg, &vpc_log_data)
	if err != nil {
		return fmt.Errorf("unable to load the logs to the struct, log %s, error: %s",
			string(vpc_log_msg), err)
	}
	slog.Debug("Generating unique string for the log")
	unique_str, err := createUniqueStrAwsVpcLog(vpc_log_data)
	slog.Debug(fmt.Sprintf("Unique string: %s", unique_str))

	if err != nil {
		return fmt.Errorf("unable to create unique string for the log, err: %s", err)
	}

	slog.Debug("Checking if the log is duplicate or not")
	value, err := rediscache.GetValue(unique_str)
	if err != nil {
		return fmt.Errorf("unable to get value from redis, key: %s, err: %s", unique_str, err)
	}
	if value == "" {
		slog.Debug("Got empty message from redis, it means log is not duplicate")
		// Non duplicate log
		slog.Debug("Sending the message to outgoing queue")

		outputchannel.OutputChannel <- vpc_log_msg

		err = rediscache.SetValue(unique_str, string(vpc_log_msg))
		slog.Debug("Setting the value in redis, So that it can be used for future deduplication")
		if err != nil {
			return fmt.Errorf("unable to set the value in redis, err: %s", err)
		}
	} else {
		slog.Debug("Log is duplicate, so not sending it to outgoing queue")
	}

	return nil
}

func createUniqueStrAwsVpcLog(vpc_log models.VpcNormalizedData) (string, error) {
	unique_string_fields := config.Config.LogSource.AwsVpcLogsModel.UniqueStringFields
	fields := strings.Split(unique_string_fields, ",")
	val := reflect.ValueOf(vpc_log)
	typ := reflect.TypeOf(vpc_log)
	unique_string := ""

	for _, field := range fields {
		field = strings.TrimSpace(field)
		// Check field exist
		_, found := typ.FieldByName(field)
		if !found {
			return "", fmt.Errorf("field '%s' does not exist in the struct", field)
		}

		// Fetch value using field name
		value := val.FieldByName(field)
		switch value.Kind() {
		case reflect.String:
			unique_string = unique_string + strings.TrimSpace(value.String())
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			unique_string = unique_string + fmt.Sprintf("%d", value.Int())
		default:
			return "", fmt.Errorf("field '%s' is not string or int", field)
		}
	}
	DEFAULT_UNIQUE_STRING := "awsvpclogs_"
	unique_string = DEFAULT_UNIQUE_STRING + unique_string
	return unique_string, nil
}
