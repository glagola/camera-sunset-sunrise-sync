package main

import (
	"net/http"
	"os"
	"time"

	"github.com/glagola/camera-sunset-sunrise-sync/internal/adapter/asecam"
	sun "github.com/glagola/camera-sunset-sunrise-sync/internal/adapter/sunrise-sunset.org"
	"github.com/glagola/camera-sunset-sunrise-sync/internal/config"
	"github.com/glagola/camera-sunset-sunrise-sync/internal/utils"
)

func main() {
	logger := utils.SetupLogger(utils.EnvProd)
	config := config.MustRead(
		utils.LoggerForPackage(logger, "config"),
		".env",
	)

	logger = utils.SetupLogger(config.Env)

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	camera, err := asecam.New(
		httpClient,
		utils.LoggerForPackage(logger, "asecam"),
		config.BaseUrl,
		config.User,
		config.HashedPassword,
	)

	if err != nil {
		os.Exit(1)
	}

	sun := sun.New(
		httpClient,
		utils.LoggerForPackage(logger, "api.sunrise-sunset.org"),
	)

	sunTimings, err := sun.GetTimings(
		config.Location.Latitude,
		config.Location.Longitude,
	)
	if err != nil {
		os.Exit(1)
	}

	if err := camera.UpdateDayTimings(sunTimings.Sunrise, sunTimings.Sunset); err != nil {
		os.Exit(1)
	}
}
