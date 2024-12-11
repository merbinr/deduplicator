package main

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/merbinr/deduplicator/internal/queue/incoming"
)

func main() {
	slog.Info("Initializing deduplication process")
	initializer()
	timeout_seconds := get_timeout_seconds()
	fmt.Println("De-duplication process started")
	for {
		err := incoming.ConsumeMessage()

		slog.Debug(fmt.Sprintf("Consumed message from incoming queue, it will be re-run after %d seconds",
			timeout_seconds))

		if err != nil {
			slog.Error(fmt.Sprintf("unable to process dedplucation, err: %s", err))
		}
		slog.Info(fmt.Sprintf("Sleeping for %d seconds", timeout_seconds))
		time.Sleep(time.Duration(timeout_seconds) * time.Second)
	}
}

func get_timeout_seconds() int {
	sleep_intervel_time_str := os.Getenv("DEDUPLICATOR_SLEEP_INTERVEL_TIME")
	slog.Debug(fmt.Sprintf("DEDUPLICATOR_SLEEP_INTERVEL_TIME value: %s", sleep_intervel_time_str))
	default_timeout_time := 2 // 2 Second
	if sleep_intervel_time_str == "" {
		slog.Debug(fmt.Sprintf("DEDUPLICATOR_SLEEP_INTERVEL_TIME value is empty, using default value: %d", default_timeout_time))
		return default_timeout_time
	} else {
		sleep_intervel, err := strconv.Atoi(sleep_intervel_time_str)
		if err != nil {
			err_msg := fmt.Sprintf("unable to convert DEDUPLICATOR_SLEEP_INTERVEL_TIME value to int value: %s, err: %s",
				sleep_intervel_time_str, err)
			slog.Error(err_msg)
			return default_timeout_time
		}
		return sleep_intervel
	}
}
