package openmeteo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"weather-cli/internal/domain"
)

const (
	geocodingURL = "https://geocoding-api.open-meteo.com/v1/search"
	forecastURL  = "https://api.open-meteo.com/v1/forecast"
)

type geocodingResponse struct {
	Results []geocodingResult `json:"results"`
}

type geocodingResult struct {
	Name      string  `json:"name"`
	Admin1    string  `json:"admin1"`
	Country   string  `json:"country"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type forecastResponse struct {
	Current forecastCurrent `json:"current"`
}

type forecastCurrent struct {
	Time                string  `json:"time"`
	Temperature2M       float64 `json:"temperature_2m"`
	ApparentTemperature float64 `json:"apparent_temperature"`
	WeatherCode         int     `json:"weather_code"`
	WindSpeed10M        float64 `json:"wind_speed_10m"`
	WindDirection10M    int     `json:"wind_direction_10m"`
	RelativeHumidity2M  int     `json:"relative_humidity_2m"`
	SurfacePressure     float64 `json:"surface_pressure"`
	Visibility          float64 `json:"visibility"`
	Precipitation       float64 `json:"precipitation"`
}

type Client struct {
	HTTPClient *http.Client
}

func NewClient() *Client {
	return &Client{
		HTTPClient: &http.Client{
			Timeout: time.Second * 5,
		},
	}
}

func (c *Client) geocode(ctx context.Context, city string) (name string, lat, lon float64, err error) {
	if city == "" {
		return "", 0, 0, fmt.Errorf("город не указан")
	}

	params := url.Values{}
	params.Set("name", city)
	params.Set("count", "1")
	params.Set("language", "ru")
	params.Set("format", "json")

	requestURL := geocodingURL + "?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
	if err != nil {
		return "", 0, 0, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", 0, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", 0, 0, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var data geocodingResponse
	if err = json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", 0, 0, err
	}

	if len(data.Results) == 0 {
		return "", 0, 0, fmt.Errorf("город не найден: %s", city)
	}

	result := data.Results[0]

	locationParts := []string{result.Name}
	if result.Admin1 != "" {
		locationParts = append(locationParts, result.Admin1)
	}
	if result.Country != "" {
		locationParts = append(locationParts, result.Country)
	}

	name = strings.Join(locationParts, ", ")

	return name, result.Latitude, result.Longitude, nil
}

func (c *Client) forecast(ctx context.Context, lat, lon float64, days int) (*forecastResponse, error) {
	params := url.Values{}
	params.Set("latitude", fmt.Sprintf("%f", lat))
	params.Set("longitude", fmt.Sprintf("%f", lon))
	params.Set("timezone", "auto")
	params.Set("wind_speed_unit", "ms")
	params.Set("days", strconv.Itoa(days))
	params.Set(
		"current",
		"temperature_2m,apparent_temperature,weather_code,"+
			"wind_speed_10m,wind_direction_10m,relative_humidity_2m,"+
			"surface_pressure,visibility,precipitation",
	)
	params.Set("hourly", "temperature_2m,precipitation_probability,windspeed_10m")
	params.Set("daily", "temperature_2m_max,temperature_2m_min,precipitation_probability_max,windspeed_10m_max,weather_code")

	requestURL := forecastURL + "?" + params.Encode()
	req, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var data forecastResponse
	if err = json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return &data, nil
}

func (c *Client) GetToday(ctx context.Context, city string) (domain.Today, error) {
	fullCity, lat, lon, err := c.geocode(ctx, city)
	if err != nil {
		return domain.Today{}, err
	}

	data, err := c.forecast(ctx, lat, lon, 1)
	if err != nil {
		return domain.Today{}, err
	}

	if data.Current.Time == "" {
		return domain.Today{}, fmt.Errorf("no current weather data")
	}

	updatedAt, err := time.Parse("2006-01-02T15:04", data.Current.Time)
	if err != nil {
		return domain.Today{}, fmt.Errorf("parse current weather time: %w", err)
	}

	return domain.Today{
		City:             fullCity,
		TemperatureC:     data.Current.Temperature2M,
		FeelsLikeC:       data.Current.ApparentTemperature,
		Condition:        weatherCodeToText(data.Current.WeatherCode),
		WindSpeedMS:      data.Current.WindSpeed10M,
		WindDirectionDeg: data.Current.WindDirection10M,
		HumidityPercent:  data.Current.RelativeHumidity2M,
		PressureHPa:      int(data.Current.SurfacePressure),
		VisibilityKm:     data.Current.Visibility / 1000,
		PrecipitationMm:  data.Current.Precipitation,
		UpdatedAt:        updatedAt,
	}, nil
}

func weatherCodeToText(code int) string {
	switch {
	case code == 0:
		return "Ясно"
	case code >= 1 && code <= 3:
		return "Переменная облачность"
	case code == 45 || code == 48:
		return "Туман"
	case code >= 51 && code <= 57:
		return "Морось"
	case code >= 61 && code <= 65:
		return "Дождь"
	case code >= 71 && code <= 77:
		return "Снег"
	case code >= 80 && code <= 82:
		return "Ливень"
	case code >= 85 && code <= 86:
		return "Снегопад"
	case code >= 95:
		return "Гроза"
	default:
		return "Неизвестно"
	}
}
