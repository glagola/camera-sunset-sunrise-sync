package config

import (
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env            string `env:"ENV" env-default:"prod"` // dev, prod
	Host           string `env:"CAMERA_HOST" env-required:"true"`
	User           string `env:"CAMERA_USER" env-default:"admin"`
	HashedPassword string `env:"CAMERA_HASHED_PASSWORD" env-required:"true"`

	Location struct {
		Latitude  float32 `env:"CAMERA_LOCATION_LATITUDE" env-required:"true"`
		Longitude float32 `env:"CAMERA_LOCATION_LONGITUDE" env-required:"true"`
	}
}

func MustRead(pathToFile string) *Config {
	logger := log.New(os.Stderr, "", log.LstdFlags)
	cfg := Config{}

	if err := cleanenv.ReadConfig(pathToFile, &cfg); err != nil {
		logger.Printf(`Failed to load config "%s" file, error "%s"`, pathToFile, err.Error())
		
		if err := cleanenv.ReadEnv(&cfg); err != nil {
			logger.Fatalf(`Failed to read environment variables error "%s"`, err.Error())
		}
	}

	return &cfg
}
