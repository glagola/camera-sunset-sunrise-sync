package asecam

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
)

type Repository struct {
	validate       *validator.Validate
	domain         string
	user           string
	hashedPassword string
}

type schedule struct {
	Hour    int `json:"hour" validate:"gte=0,lt=24"`
	Minute  int `json:"minute" validate:"gte=0,lt=60"`
	Second  int `json:"second" validate:"gte=0,lt=60"`
	Reserve int `json:"reserve"`
}

type imageSettings struct {
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

func (s *schedule) Set(new time.Time) {
	s.Hour = new.Hour()
	s.Minute = new.Minute()
}

func New(validate *validator.Validate, domain, user, hashedPassword string) *Repository {
	return &Repository{
		validate:       validate,
		domain:         domain,
		user:           user,
		hashedPassword: hashedPassword,
	}
}

func (s *Repository) buildUrl(params map[string]string) string {
	query := url.Values{}
	for k, v := range params {
		query.Add(k, v)
	}

	query.Add("username", s.user)
	query.Add("password", s.hashedPassword)

	_url := url.URL{
		Scheme:   "http",
		Host:     s.domain,
		Path:     "/cgi-bin/web.cgi",
		RawQuery: query.Encode(),
	}

	return _url.String()
}

func (s *Repository) getImageSettings() (*imageSettings, error) {
	url := s.buildUrl(map[string]string{
		"action": "get",
		"cmd":    "image",
	})

	response, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("unable to get image settings: %w", err)
	}
	defer response.Body.Close()

	result := imageSettings{}
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
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

func (s *Repository) setImageSettings(imageSettings imageSettings) error {

	str, err := json.Marshal(imageSettings)
	if err != nil {
		return err
	}

	url := s.buildUrl(map[string]string{
		"action": "set",
		"cmd":    "image",
		"param":  string(str),
	})

	response, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("unable to set image settings: %w", err)
	}
	defer response.Body.Close()

	var body *struct {
		Status string
		Data   string
	}

	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		return fmt.Errorf("failed to set image settings: %w", err)
	}

	if body == nil || body.Status != "ok" {
		return fmt.Errorf("failed to set image settings")
	}

	return nil
}

type timezones map[int]*time.Location

var timezoneOptionRegexp regexp.Regexp = *regexp.MustCompile(`(?im)<\s*option\s*value\s*=\s*"(\d+)"\s*>\s*(UTC([+-])(\d+):(\d+))\s*</\s*option\s*>`)

func (s *Repository) getTimezones() (timezones, error) {
	result := make(timezones, 34)

	url := (&url.URL{
		Scheme: "http",
		Host:   s.domain,
		Path:   "/view/time_setting.html",
	}).String()

	response, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("unable to make request to %s: %w", url, err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read response from %s: %w", url, err)
	}

	matches := timezoneOptionRegexp.FindAllStringSubmatch(string(body), -1)

	var timezoneId, hour, minute int
	for _, match := range matches {
		sign := 1
		if match[3] == "-" {
			sign = -1
		}

		if timezoneId, err = strconv.Atoi(match[1]); err != nil {
			return nil, fmt.Errorf("unable to parse timezone id: %w", err)
		}

		if hour, err = strconv.Atoi(match[4]); err != nil {
			return nil, fmt.Errorf("unable to parse timezone hour: %w", err)
		}

		if minute, err = strconv.Atoi(match[5]); err != nil {
			return nil, fmt.Errorf("unable to parse timezone minute: %w", err)
		}

		result[timezoneId] = time.FixedZone(match[2], sign*(minute*60+hour*60*60))
	}

	return result, nil
}

func (s *Repository) getTimezone() (*time.Location, error) {
	timezoneById, err := s.getTimezones()
	if err != nil {
		return nil, fmt.Errorf("failed to get timezones list: %w", err)
	}

	id, err := s.getTimezoneId()
	if err != nil {
		return nil, fmt.Errorf("failed to get current timezone id: %w", err)
	}

	timezone, exists := timezoneById[id]
	if !exists {
		return nil, fmt.Errorf("failed to get timezone offset by id")
	}

	return timezone, nil
}

func (s *Repository) getTimezoneId() (int, error) {
	url := s.buildUrl(map[string]string{
		"action": "get",
		"cmd":    "systime",
	})

	response, err := http.Get(url)
	if err != nil {
		return 0, fmt.Errorf("failed to make GET request to %s: %w", url, err)
	}
	defer response.Body.Close()

	var systemTimeSettings *struct {
		Timezone int `json:"timezone"`
		TimeSec  int `json:"time_sec"`
	}

	if err := json.NewDecoder(response.Body).Decode(&systemTimeSettings); err != nil {
		return 0, fmt.Errorf("unable to parse response: %w", err)
	}

	if systemTimeSettings == nil {
		return 0, fmt.Errorf("invalid response")
	}

	return systemTimeSettings.Timezone, nil
}

func (s *Repository) UpdateDayTimings(logger *slog.Logger, sunrise, sunset time.Time) error {
	cameraTimezone, err := s.getTimezone()
	if err != nil {
		logger.Error("Failed to get current camera timezone", slog.Any("error", err))
		return fmt.Errorf("failed to get current timezone: %w", err)
	}

	imageSettings, err := s.getImageSettings()
	if err != nil {
		logger.Error("Failed to get asecam image settings", slog.Any("error", err))
		return fmt.Errorf("failed to get asecam image settings: %w", err)
	}

	imageSettings.DayBegin.Set(sunrise.In(cameraTimezone))
	imageSettings.DayEnd.Set(sunset.In(cameraTimezone))

	if err := s.setImageSettings(*imageSettings); err != nil {
		logger.Error("Failed to set updated image settings", slog.Any("error", err))
		return fmt.Errorf("failed to set updated image settings: %w", err)
	}

	return nil
}
