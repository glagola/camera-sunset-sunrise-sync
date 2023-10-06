package main

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/glagola/asecam-day-light-updater/internal/adapter/asecam"
	sun "github.com/glagola/asecam-day-light-updater/internal/adapter/sunrise-sunset.org"
	"github.com/glagola/asecam-day-light-updater/internal/config"
	"github.com/glagola/asecam-day-light-updater/internal/logger"
)

func main() {
	config := config.MustRead(".env")

	logger := logger.SetupLogger(config.Env)

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	camera := asecam.New(
		httpClient,
		logger.With(slog.String("adapter", "asecam")),
		config.Host,
		config.User,
		config.HashedPassword,
	)

	sun := sun.New(
		httpClient, 
		logger.With(slog.String("adapter", "api.sunrise-sunset.org")),
	)

	sunTimings, err := sun.GetTimings(
		config.Location.Latitude,
		config.Location.Longitude,
	)
	if err != nil {
		panic(err)
	}

	if err := camera.UpdateDayTimings(sunTimings.Sunrise, sunTimings.Sunset); err != nil {
		panic(err)
	}
}