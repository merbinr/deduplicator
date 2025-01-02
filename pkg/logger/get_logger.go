package logger

import (
	"log/slog"
	"os"
	"sync"

	"github.com/natefinch/lumberjack"
)

var (
	instance *slog.Logger
	once     sync.Once
)

func GetLogger() *slog.Logger {
	once.Do(func() {
		logFile := &lumberjack.Logger{
			Filename:   "./app_logs/app.log", // Path to your log file
			MaxSize:    10,                   // Max size in MB before rotation
			MaxAge:     5,                    // Maximum number of days to retain logs
			MaxBackups: 0,                    // No limit on the number of backups
			Compress:   true,                 // Compress old log files
		}

		handler := slog.NewTextHandler(logFile, &slog.HandlerOptions{
			Level: getLogLevel(),
		})

		instance = slog.New(handler)
	})
	return instance
}

func getLogLevel() slog.Level {
	logLevel := os.Getenv("CATCHER_LOG_LEVEL")
	if logLevel == "" {
		return slog.LevelInfo
	}
	switch logLevel {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
