package main

type Config struct {
	Host           string `env:"CAMERA_HOST"`
	User           string `env:"CAMERA_USER"`
	HashedPassword string `env:"CAMERA_HASHED_PASSWORD"`

	Location struct {
		Latitude  float32 `env:"CAMERA_LOCATION_LATITUDE"`
		Longitude float32 `env:"CAMERA_LOCATION_LONGITUDE"`
	}
}
