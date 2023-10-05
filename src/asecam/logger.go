package asecam

import (
	"log/slog"
	"net/http"
	"time"
)

type scheduleLogRecord struct {
	time *schedule
	location *time.Location
}

func (s scheduleLogRecord) LogValue() slog.Value {
	return slog.AnyValue(time.Date(0, 0, 0, s.time.Hour, s.time.Minute, s.time.Second, 0, s.location))
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