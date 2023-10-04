package main

import (
	"asecam/src/asecam"
	"asecam/src/sun"
	"log/slog"
	"os"

	"github.com/go-playground/validator/v10"
)

func main() {
	logger := slog.New(
		slog.NewTextHandler(os.Stderr, nil),
	)

	config := MustReadConfig(logger, ".env")

	validator := validator.New(validator.WithRequiredStructEnabled())

	camera := asecam.NewAdapter(
		validator,
		config.Host,
		config.User,
		config.HashedPassword,
	)

	sunRepo := sun.NewRepository(validator)

	sunTimings, err := sunRepo.GetTimings(
		config.Location.Latitude,
		config.Location.Longitude,
	)
	if err != nil {
		panic(err)
	}

	if err := camera.UpdateSunTimings(logger, sunTimings.Sunrise, sunTimings.Sunset); err != nil {
		panic(err)
	}
}
