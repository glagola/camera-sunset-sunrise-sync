package sun

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/go-playground/validator/v10"
)

type SunRepository struct {
	validate *validator.Validate
}

type SunTimings struct {
	Sunrise                   time.Time `json:"sunrise"`
	Sunset                    time.Time `json:"sunset"`
	SolarNoon                 time.Time `json:"solar_noon"`
	DayLength                 int       `json:"day_length"`
	CivilTwilightBegin        time.Time `json:"civil_twilight_begin"`
	CivilTwilightEnd          time.Time `json:"civil_twilight_end"`
	NauticalTwilightBegin     time.Time `json:"nautical_twilight_begin"`
	NauticalTwilightEnd       time.Time `json:"nautical_twilight_end"`
	AstronomicalTwilightBegin time.Time `json:"astronomical_twilight_begin"`
	AstronomicalTwilightEnd   time.Time `json:"astronomical_twilight_end"`
}

type sunriseSunsetResponse struct {
	Results *SunTimings `json:"results" validate:"required"`
	Status  string      `json:"status"`
}

func NewRepository(validate *validator.Validate) SunRepository {
	return SunRepository{
		validate: validate,
	}
}

func (s SunRepository) buildUrl(latitude, longitude float32) string {
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

func (s SunRepository) GetSunTimings(latitude, longitude float32) (*SunTimings, error) {
	url := s.buildUrl(latitude, longitude)

	response, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("unable to make request to %s: %w", url, err)
	}
	defer response.Body.Close()

	result := sunriseSunsetResponse{}
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("unable to unmarshal json response: %w", err)
	}

	err = s.validate.Struct(result)
	if err != nil {

		// this check is only needed when your code could produce
		// an invalid value for validation such as interface with nil
		// value most including myself do not usually have code like this.
		if _, ok := err.(*validator.InvalidValidationError); ok {
			fmt.Println(err)
			return nil, err
		}

		for _, err := range err.(validator.ValidationErrors) {

			fmt.Println(err.Namespace())
			fmt.Println(err.Field())
			fmt.Println(err.StructNamespace())
			fmt.Println(err.StructField())
			fmt.Println(err.Tag())
			fmt.Println(err.ActualTag())
			fmt.Println(err.Kind())
			fmt.Println(err.Type())
			fmt.Println(err.Value())
			fmt.Println(err.Param())

			fmt.Println()
		}

		return nil, err
	}

	return result.Results, nil
}
