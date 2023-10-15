package config

import (
	"log/slog"
	"os"

	"github.com/glagola/camera-sunset-sunrise-sync/internal/utils"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env            string `env:"ENV" env-default:"prod"` // dev, prod
	BaseUrl        string `env:"CAMERA_BASE_URL" env-required:"true"`
	User           string `env:"CAMERA_USER" env-default:"admin"`
	HashedPassword string `env:"CAMERA_HASHED_PASSWORD" env-required:"true"`

	Location struct {
		Latitude  float32 `env:"CAMERA_LOCATION_LATITUDE" env-required:"true"`
		Longitude float32 `env:"CAMERA_LOCATION_LONGITUDE" env-required:"true"`
	}
}

func MustRead(logger *slog.Logger, pathToFile string) *Config {
	logger = utils.LoggerForMethod(logger, "MustRead")

	cfg := Config{}

	if err := cleanenv.ReadConfig(pathToFile, &cfg); err != nil {
		logger.Warn("Failed to load config from file", slog.String("file", pathToFile), slog.Any("error", err.Error()))

		if err := cleanenv.ReadEnv(&cfg); err != nil {
			logger.Error("Failed to load config from environment variables", slog.Any("error", err.Error()))
			os.Exit(1)
		}
	}

	return &cfg
}
