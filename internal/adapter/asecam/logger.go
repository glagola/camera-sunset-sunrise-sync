package asecam

import (
	"log/slog"
	"time"
)

type scheduleLogRecord struct {
	time     *schedule
	location *time.Location
}

func (s scheduleLogRecord) LogValue() slog.Value {
	return slog.AnyValue(time.Date(0, 0, 0, s.time.Hour, s.time.Minute, s.time.Second, 0, s.location))
}
