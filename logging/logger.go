package logging

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/PRPO-skupina-02/common/config"
)

func GetDefaultLogger() *slog.Logger {

	var logger *slog.Logger

	logLevelConfig := config.GetEnvDefault("LOG_LEVEL", "INFO")
	var logLevel = new(slog.LevelVar)
	switch logLevelConfig {
	case "DEBUG":
		logLevel.Set(slog.LevelDebug)
	case "INFO":
		logLevel.Set(slog.LevelInfo)
	case "ERROR":
		logLevel.Set(slog.LevelError)
	default:
		logLevel.Set(slog.LevelInfo)
	}

	slog.Info(fmt.Sprintf("Log level: %s", logLevel.Level()))

	logHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})
	logger = slog.New(logHandler)

	return logger
}
