package sun

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/glagola/camera-sunset-sunrise-sync/internal/utils"
)

type Adapter struct {
	client *http.Client
	logger *slog.Logger
}

type SunTimings struct {
	Sunrise time.Time `json:"sunrise"`
	Sunset  time.Time `json:"sunset"`
}

func New(client *http.Client, logger *slog.Logger) Adapter {
	return Adapter{
		client: client,
		logger: logger,
	}
}

func (s Adapter) buildUrl(latitude, longitude float32) string {
	values := url.Values{}

	values.Add("lat", fmt.Sprintf("%f", latitude))
	values.Add("lng", fmt.Sprintf("%f", longitude))
	values.Add("formatted", "0")

	return (&url.URL{
		Scheme:   "https",
		Host:     "api.sunrise-sunset.org",
		Path:     "json",
		RawQuery: values.Encode(),
	}).String()
}

func (s Adapter) GetTimings(latitude, longitude float32) (*SunTimings, error) {
	logger := s.logger.With(slog.String("method", "GetTimings"))
	logger.Debug(
		"Get sunrise and sunset for the location",
		slog.Group("location",
			slog.Float64("latitude", float64(latitude)),
			slog.Float64("longitude", float64(longitude)),
		),
	)

	url := s.buildUrl(latitude, longitude)

	response, err := s.client.Get(url)
	if err != nil {
		utils.LogHttpError(logger, "Failed to get sun timings", url, response)
		return nil, fmt.Errorf("unable to make request to %s: %w", url, err)
	}
	defer response.Body.Close()

	var result struct {
		Results *SunTimings `json:"results"`
		Status  string      `json:"status"`
	}

	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		logger.Error("Failed to parse json")
		return nil, fmt.Errorf("unable to unmarshal json response: %w", err)
	}

	if result.Results == nil {
		logger.Error("Results are empty")
		return nil, fmt.Errorf("no data in response")
	}

	return result.Results, nil
}
