package main

import (
	"asecam/config"
	"asecam/logger"
	"asecam/src/asecam"
	"asecam/src/sun"
	"net/http"
	"time"
)

func main() {
	config := config.MustRead(".env")

	logger := logger.SetupLogger(config.Env)

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