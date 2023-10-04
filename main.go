package main

import (
	"asecam/src/asecam"
	"asecam/src/sun"
	"log/slog"
	"net/http"
	"os"
	"time"
)

func main() {
	logger := slog.New(
		slog.NewTextHandler(os.Stderr, nil),
	)
	slog.SetDefault(logger)

	config := MustReadConfig(logger, ".env")

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	camera := asecam.New(
		httpClient,
		config.Host,
		config.User,
		config.HashedPassword,
	)

	sun := sun.New(httpClient)

	sunTimings, err := sun.GetTimings(
		config.Location.Latitude,
		config.Location.Longitude,
	)
	if err != nil {
		panic(err)
	}

	if err := camera.UpdateDayTimings(logger, sunTimings.Sunrise, sunTimings.Sunset); err != nil {
		panic(err)
	}
}
