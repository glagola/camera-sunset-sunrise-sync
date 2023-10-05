package logger

import (
	"log/slog"
	"os"
)

const (
	EnvDev = "dev"
	EnvProd = "prod"
)

func SetupLogger(requestedEnv string) *slog.Logger {
	var handler slog.Handler
	env := requestedEnv

	switch requestedEnv {
	case EnvDev:
		handler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})

	default:
		env = EnvProd
		fallthrough
	case EnvProd:
		handler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	}

	logger := slog.New(handler)

	logger.Info("Logger init environment", slog.String("loaded", env), slog.String("requested", requestedEnv))

	slog.SetDefault(logger)

	return logger
}