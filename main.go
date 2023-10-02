package main

import (
	"asecam/src/asecam"
	"asecam/src/sun"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/ilyakaznacheev/cleanenv"
)

func main() {
	var config Config

	if err := cleanenv.ReadConfig(".env", &config); err != nil {
		fmt.Printf("Failed to load .env config: %e", err)
		panic(err)
	}

	validator := validator.New(validator.WithRequiredStructEnabled())

	asecamRepo := asecam.NewRepository(
		validator,
		config.Host,
		config.User,
		config.HashedPassword,
	)

	sunRepo := sun.NewRepository(validator)

	sunTimings, err := sunRepo.GetSunTimings(
		config.Location.Latitude,
		config.Location.Longitude,
	)
	if err != nil {
		fmt.Printf("Failed to get sun timings: %e", err)
		panic(err)
	}

	cameraTimezone, err := asecamRepo.GetTimezone()
	if err != nil {
		fmt.Printf("Failed to get current timezone: %e", err)
		panic(err)
	}

	imageSettings, err := asecamRepo.GetImageSettings()
	if err != nil {
		fmt.Printf("Failed to get asecam image settings: %e", err)
		panic(err)
	}

	sunrise := sunTimings.Sunrise.In(cameraTimezone)
	sunset := sunTimings.Sunset.In(cameraTimezone)

	fmt.Printf("New sunrise %s, in target TZ %s\n", sunTimings.Sunrise, sunrise)
	fmt.Printf("New sunrise %s, in target TZ %s\n", sunTimings.Sunset, sunset)

	imageSettings.DayBegin.Set(sunrise)
	imageSettings.DayEnd.Set(sunset)

	if err := asecamRepo.SetImageSettings(*imageSettings); err != nil {
		fmt.Printf("Failed to set updated image settings: %e", err)
		panic(err)
	}
}
