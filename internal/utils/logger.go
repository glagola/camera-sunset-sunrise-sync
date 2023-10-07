package utils

import (
	"log/slog"
	"net/http"
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

	logger.Info(
		"Logger init", 
		slog.Group(
			"environment", 
			slog.String("loaded", env), 
			slog.String("requested", requestedEnv),
		),
	)

	slog.SetDefault(logger)

	return logger
}


func LogHttpError(logger *slog.Logger, message string, url string, response *http.Response) {
	logger.Error(
		message,
		slog.String("url", url),
		slog.Group(
			"response",
			slog.String("msg", response.Status),
			slog.Int("code", response.StatusCode),
		),
	)
}

func LoggerForMethod(logger *slog.Logger, method string) *slog.Logger {
	return logger.With(slog.String("method", method))
}

func LoggerForPackage(logger *slog.Logger, _package string) *slog.Logger {
	return logger.With(slog.String("package", _package))
}