package main

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/merbinr/deduplicator/internal/deduplication"
)

func main() {
	initializer()
	timeout_seconds := get_timeout_seconds()

	for {
		fmt.Println("Running...")
		time.Sleep(time.Duration(timeout_seconds))
		err := deduplication.ProcessDeduplication()
		if err != nil {
			slog.Error(fmt.Sprintf("unable to process dedplucation, err: %s", err))
		}
	}

}

func get_timeout_seconds() int {
	sleep_intervel_time_str := os.Getenv("DEDUPLICATOR_SLEEP_INTERVEL_TIME")
	default_timeout_time := 10 // 10 Seconds
	if sleep_intervel_time_str == "" {
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
