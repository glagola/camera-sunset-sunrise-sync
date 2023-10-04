package sun

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type Adapter struct {
	client *http.Client
}

type SunTimings struct {
	Sunrise time.Time `json:"sunrise"`
	Sunset  time.Time `json:"sunset"`
}

func New(client *http.Client) Adapter {
	return Adapter{
		client: client,
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
	url := s.buildUrl(latitude, longitude)

	response, err := s.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("unable to make request to %s: %w", url, err)
	}
	defer response.Body.Close()

	var result struct {
		Results *SunTimings `json:"results"`
		Status  string      `json:"status"`
	}

	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("unable to unmarshal json response: %w", err)
	}

	if result.Results == nil {
		return nil, fmt.Errorf("no data in response")
	}

	return result.Results, nil
}
