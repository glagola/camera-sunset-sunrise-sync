package main

import (
	"log/slog"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Host           string `env:"CAMERA_HOST"`
	User           string `env:"CAMERA_USER"`
	HashedPassword string `env:"CAMERA_HASHED_PASSWORD"`

	Location struct {
		Latitude  float32 `env:"CAMERA_LOCATION_LATITUDE"`
		Longitude float32 `env:"CAMERA_LOCATION_LONGITUDE"`
	}
}

func MustReadConfig(logger *slog.Logger, pathToFile string) *Config {
	cfg := Config{}

	if err := cleanenv.ReadConfig(pathToFile, &cfg); err != nil {
		logger.Error("Failed to load config", slog.Any("error", err))
		os.Exit(1)
	}

	return &cfg
}
