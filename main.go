package main

import (
	"asecam/src/asecam"
	"asecam/src/sun"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-playground/validator/v10"
)

func main() {
	logger := slog.New(
		slog.NewTextHandler(os.Stderr, nil),
	)
	slog.SetDefault(logger)

	config := MustReadConfig(logger, ".env")

	validator := validator.New(validator.WithRequiredStructEnabled())

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	camera := asecam.New(
		validator,
		httpClient,
		config.Host,
		config.User,
		config.HashedPassword,
	)

	sun := sun.New(validator, httpClient)

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
