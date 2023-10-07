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
	"strings"
	"time"

	"github.com/glagola/camera-sunset-sunrise-sync/internal/utils"
)

type Adapter struct {
	client         *http.Client
	logger         *slog.Logger
	baseUrl        *url.URL
	user           string
	hashedPassword string
}

type schedule struct {
	Hour    int `json:"hour"`
	Minute  int `json:"minute"`
	Second  int `json:"second"`
	Reserve int `json:"reserve"`
}

type imageSettings struct {
	Brightness           int       `json:"brightness"`
	Saturation           int       `json:"saturation"`
	Contrast             int       `json:"contrast"`
	Sharpness            int       `json:"sharpness"`
	Exposure             int       `json:"exposure"`
	DayNightMode         int       `json:"day_night_mode"`
	DayBegin             *schedule `json:"day_begin"`
	DayEnd               *schedule `json:"day_end"`
	Mirror               int       `json:"mirror"`
	Flip                 int       `json:"flip"`
	WdrEnable            int       `json:"wdr_enable"`
	IRCutDelay           int       `json:"ircut_delay"`
	AntiFlickerEnable    int       `json:"anti_flicker_enable"`
	BacklightEnable      int       `json:"backlight_enable"`
	TvStandard           int       `json:"tv_standard"`
	DrcStrength          int       `json:"drc_strenght"`
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

func New(client *http.Client, logger *slog.Logger, baseUrl, user, hashedPassword string) (*Adapter, error) {
	_url, err := url.Parse(baseUrl)

	if err != nil {
		logger.Error(
			"Invalid baseUrl",
			slog.String("baseUrl", baseUrl),
			slog.String("err", err.Error()),
		)
		return nil, fmt.Errorf(`invalid baseUrl = "%s"`, baseUrl)
	}

	if _url.Host == "" {
		logger.Error("BaseUrl must have domain/ip address", slog.String("baseUrl", baseUrl))
		return nil, fmt.Errorf("baseUrl must have domain/ip address")
	}

	_url.Path = strings.TrimRight(_url.Path, "/")

	return &Adapter{
		client:         client,
		logger:         logger,
		baseUrl:        _url,
		user:           user,
		hashedPassword: hashedPassword,
	}, nil
}

func (s Adapter) buildUrl(params map[string]string) string {
	query := url.Values{}
	for k, v := range params {
		query.Add(k, v)
	}

	query.Add("username", s.user)
	query.Add("password", s.hashedPassword)

	_url := url.URL{
		Scheme:   s.baseUrl.Scheme,
		Host:     s.baseUrl.Host,
		Path:     fmt.Sprintf("%s/cgi-bin/web.cgi", s.baseUrl.Path),
		RawQuery: query.Encode(),
	}

	return _url.String()
}

func (s Adapter) getImageSettings() (*imageSettings, error) {
	logger := utils.LoggerForMethod(s.logger, "getImageSettings")

	logger.Debug("Get camera's image settings")

	url := s.buildUrl(map[string]string{
		"action": "get",
		"cmd":    "image",
	})

	response, err := s.client.Get(url)
	if err != nil {
		utils.LogHttpError(logger, "Failed to get image settings", url, response)
		return nil, fmt.Errorf("unable to get image settings: %w", err)
	}
	defer response.Body.Close()

	result := imageSettings{}
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		logger.Error("Failed to parse json")
		return nil, fmt.Errorf("unable to response: %w", err)
	}

	return &result, nil
}

func (s Adapter) setImageSettings(imageSettings imageSettings) error {
	logger := utils.LoggerForMethod(s.logger, "setImageSettings")

	logger.Debug(
		"Set image settings",
		slog.Group(
			"params",
			slog.Any("imageSettings", imageSettings),
		),
	)

	str, err := json.Marshal(imageSettings)
	if err != nil {
		return err
	}

	url := s.buildUrl(map[string]string{
		"action": "set",
		"cmd":    "image",
		"param":  string(str),
	})

	response, err := s.client.Get(url)
	if err != nil {
		utils.LogHttpError(logger, "Failed to set image settings", url, response)
		return fmt.Errorf("unable to set image settings: %w", err)
	}
	defer response.Body.Close()

	var body *struct {
		Status string
		Data   string
	}

	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		logger.Error("Failed to parse json")
		return fmt.Errorf("failed to set image settings: %w", err)
	}

	if body == nil || body.Status != "ok" {
		logger.Error("Refused to change image settings")
		return fmt.Errorf("failed to set image settings")
	}

	return nil
}

type timezones map[int]*time.Location

var timezoneOptionRegexp regexp.Regexp = *regexp.MustCompile(`(?im)<\s*option\s*value\s*=\s*"(\d+)"\s*>\s*(UTC([+-])(\d+):(\d+))\s*</\s*option\s*>`)

func (s Adapter) getTimezones() (timezones, error) {
	logger := utils.LoggerForMethod(s.logger, "getTimezones")

	logger.Debug("Get camera's timezones list")

	result := make(timezones, 34)

	url := (&url.URL{
		Scheme: s.baseUrl.Scheme,
		Host:   s.baseUrl.Host,
		Path:   fmt.Sprintf("%s/view/time_setting.html", s.baseUrl.Path),
	}).String()

	response, err := s.client.Get(url)
	if err != nil {
		utils.LogHttpError(logger, "Failed to get timezones", url, response)
		return nil, fmt.Errorf("unable to make request to %s: %w", url, err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		logger.Error("Failed to read response body")
		return nil, fmt.Errorf("unable to read response from %s: %w", url, err)
	}

	logger.Debug("Html with timezones fetched", slog.Int("bytes", len(body)))

	matches := timezoneOptionRegexp.FindAllStringSubmatch(string(body), -1)

	var timezoneId, hour, minute int
	for _, match := range matches {
		sign := 1
		if match[3] == "-" {
			sign = -1
		}

		if timezoneId, err = strconv.Atoi(match[1]); err != nil {
			logger.Error("Unable to parse timezone id", slog.String("str", match[1]))
			return nil, fmt.Errorf("unable to parse timezone id: %w", err)
		}

		if hour, err = strconv.Atoi(match[4]); err != nil {
			logger.Error("Unable to parse timezone hour", slog.String("str", match[4]))
			return nil, fmt.Errorf("unable to parse timezone hour: %w", err)
		}

		if minute, err = strconv.Atoi(match[5]); err != nil {
			logger.Error("Unable to parse timezone minute", slog.String("str", match[5]))
			return nil, fmt.Errorf("unable to parse timezone minute: %w", err)
		}

		result[timezoneId] = time.FixedZone(match[2], sign*(minute*60+hour*60*60))
	}

	return result, nil
}

func (s Adapter) getTimezone() (*time.Location, error) {
	logger := utils.LoggerForMethod(s.logger, "getTimezone")

	logger.Debug("Get camera's current timezone offset")

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
		logger.Error(
			"The timezone is unknown",
			slog.Int("id", id),
		)
		return nil, fmt.Errorf("failed to get timezone offset by id")
	}

	return timezone, nil
}

func (s Adapter) getTimezoneId() (int, error) {
	logger := utils.LoggerForMethod(s.logger, "getTimezoneId")

	logger.Debug("Get timezone id")

	url := s.buildUrl(map[string]string{
		"action": "get",
		"cmd":    "systime",
	})

	response, err := s.client.Get(url)
	if err != nil {
		utils.LogHttpError(logger, "Failed to get timezone Id", url, response)
		return 0, fmt.Errorf("failed to make GET request to %s: %w", url, err)
	}
	defer response.Body.Close()

	var systemTimeSettings *struct {
		Timezone int `json:"timezone"`
		TimeSec  int `json:"time_sec"`
	}

	if err := json.NewDecoder(response.Body).Decode(&systemTimeSettings); err != nil {
		logger.Error("Failed to parse json")
		return 0, fmt.Errorf("unable to parse response: %w", err)
	}

	if systemTimeSettings == nil {
		logger.Error("No timezone info in response")
		return 0, fmt.Errorf("invalid response")
	}

	logger.Debug(
		"Fetched timezone id",
		slog.Int("timezoneId", systemTimeSettings.Timezone),
	)

	return systemTimeSettings.Timezone, nil
}

func (s Adapter) UpdateDayTimings(sunrise, sunset time.Time) error {
	logger := utils.LoggerForMethod(s.logger, "UpdateDayTimings")

	logger.Debug(
		"Update time of day light",
		slog.Group(
			"params",
			slog.Any("sunrise", sunrise),
			slog.Any("sunset", sunset),
		),
	)

	timezone, err := s.getTimezone()
	if err != nil {
		logger.Error("Failed to get camera's timezone")
		return fmt.Errorf("failed to get camera's timezone: %w", err)
	}

	logger.Info("Camera's timezone fetched")
	logger.Debug("Camera's timezone", slog.Any("timezone", timezone))

	imageSettings, err := s.getImageSettings()
	if err != nil {
		logger.Error("Failed to get camera's image settings")
		return fmt.Errorf("failed to get camera's image settings: %w", err)
	}

	logger.Info("Camera's day light settings fetched")
	logger.Debug(
		"Camera's day light settings fetched",
		slog.Any("sunrise", scheduleLogRecord{imageSettings.DayBegin, timezone}),
		slog.Any("sunset", scheduleLogRecord{imageSettings.DayEnd, timezone}),
	)

	imageSettings.DayBegin.Set(sunrise.In(timezone))
	imageSettings.DayEnd.Set(sunset.In(timezone))

	if err := s.setImageSettings(*imageSettings); err != nil {
		logger.Error("Failed to update camera's image settings")
		return fmt.Errorf("failed to update camera's image settings: %w", err)
	}

	logger.Info("Camera's time of day light updated")

	return nil
}
