package utils

import (
	"log/slog"
	"net/http"
)
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