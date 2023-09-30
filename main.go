package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/go-playground/validator/v10"
)

type asecamAction string

const (
	set asecamAction = "set"
	get asecamAction = "get"
)

type AsecamRepository struct {
	validate *validator.Validate
	domain   string
	user     string
	password string
}

type schedule struct {
	Hour    int `json:"hour" validate:"gte=0,lt=24"`
	Minute  int `json:"minute" validate:"gte=0,lt=60"`
	Second  int `json:"second" validate:"gte=0,lt=60"`
	Reserve int `json:"reserve"`
}

type AsecamImageSettings struct {
	Brightness           int       `json:"brightness"`
	Saturation           int       `json:"saturation"`
	Contrast             int       `json:"contrast"`
	Sharpness            int       `json:"sharpness"`
	Exposure             int       `json:"exposure"`
	DayNightMode         int       `json:"day_night_mode" validate:"gte=0,lt=4"`
	DayBegin             *schedule `json:"day_begin" validate:"required"`
	DayEnd               *schedule `json:"day_end" validate:"required"`
	Mirror               int       `json:"mirror"`
	Flip                 int       `json:"flip"`
	WdrEnable            int       `json:"wdr_enable"`
	IrcutDelay           int       `json:"ircut_delay"`
	AntiFlickerEnable    int       `json:"anti_flicker_enable"`
	BacklightEnable      int       `json:"backlight_enable"`
	TvStandard           int       `json:"tv_standard"`
	DrcStrenght          int       `json:"drc_strenght"`
	NrEnable             int       `json:"nr_enable"`
	LedBrightness        int       `json:"led_brightness"`
	MaxLedBrightness     int       `json:"max_led_brightness"`
	LedBrightnessMode    int       `json:"led_brightness_mode"`
	DayNightLux          int       `json:"day_night_lux"`
	FaceMode             int       `json:"face_mode"`
	SmartFaceMode        int       `json:"smart_face_mode"`
	NightFpsSelect       int       `json:"night_fps_select"`
	LdcEnable            int       `json:"ldc_enable"`
	Rotation             int       `json:"rotation"`
	DayToNightBrightness int       `json:"day_to_night_brightness"`
	NightToDayBrightness int       `json:"night_to_day_brightness"`
}

func NewAsecamRepository(validate *validator.Validate, domain, user, password string) *AsecamRepository {
	return &AsecamRepository{
		validate: validate,
		domain:   domain,
		user:     user,
		password: password,
	}
}

func (s *AsecamRepository) buildUrl(params map[string]string) string {
	query := url.Values{}
	for k, v := range params {
		query.Add(k, v)
	}

	query.Add("username", s.user)
	query.Add("password", s.password)

	_url := url.URL{
		Scheme:   "http",
		Host:     s.domain,
		Path:     "/cgi-bin/web.cgi",
		RawQuery: query.Encode(),
	}

	return _url.String()
}

func (s *AsecamRepository) GetImageSettings() (*AsecamImageSettings, error) {
	url := s.buildUrl(map[string]string{
		"action": string(get),
		"cmd":    "image",
	})

	response, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("unable to get image settings: %w", err)
	}

	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to get image settings: %w", err)
	}

	result := AsecamImageSettings{}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unable to response: %w", err)
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

		// from here you can create your own error messages in whatever language you wish
		return nil, err
	}

	return &result, nil
}

func (s *AsecamRepository) SetImageSettings(imageSettings AsecamImageSettings) error {

	str, err := json.Marshal(imageSettings)
	if err != nil {
		return err
	}

	url := s.buildUrl(map[string]string{
		"action": string(set),
		"cmd":    "image",
		"param":  string(str),
	})

	response, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("unable to set image settings: %w", err)
	}

	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("unable to set image settings: %w", err)
	}

	_ = body

	return nil
}

///////////////////

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

func NewSunRepository(validate *validator.Validate) SunRepository {
	return SunRepository{
		validate: validate,
	}
}

func (s SunRepository) buildUrl(latitude, longitude float64) (string, *url.URL) {
	values := url.Values{}

	values.Add("lat", fmt.Sprintf("%f", latitude))
	values.Add("lng", fmt.Sprintf("%f", longitude))
	values.Add("formatted", "0")

	_url := url.URL{
		Scheme:   "https",
		Host:     "api.sunrise-sunset.org",
		Path:     "json",
		RawQuery: values.Encode(),
	}

	return _url.String(), &_url
}

func (s SunRepository) GetSunTimints(latitude, longitude float64) (*SunTimings, error) {
	url, urlInfo := s.buildUrl(latitude, longitude)

	response, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("unable to make request to %s: %w", url, err)
	}

	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read response from %s: %w", urlInfo.Host, err)
	}

	result := sunriseSunsetResponse{}

	if err := json.Unmarshal(body, &result); err != nil {
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

///////////////////

func main() {

	validator := validator.New(validator.WithRequiredStructEnabled())

}
