package openmeteo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
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

func geocode(city string) (name string, lat, lon float64, err error) {
	if city == "" {
		return "", 0, 0, fmt.Errorf("город не указан")
	}

	params := url.Values{}
	params.Set("name", city)
	params.Set("count", "1")
	params.Set("language", "ru")
	params.Set("format", "json")

	requestURL := geocodingURL + "?" + params.Encode()

	resp, err := http.Get(requestURL)
	if err != nil {
		return "", 0, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", 0, 0, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var data geocodingResponse
	json.NewDecoder(resp.Body).Decode(&data)

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

func getCurrentWeather(lat, lon float64) (domain.Today, error) {
	params := url.Values{}
	params.Set("latitude", fmt.Sprintf("%f", lat))
	params.Set("longitude", fmt.Sprintf("%f", lon))
	params.Set("timezone", "auto")
	params.Set("wind_speed_unit", "ms")
	params.Set(
		"current",
		"temperature_2m,apparent_temperature,weather_code,"+
			"wind_speed_10m,wind_direction_10m,relative_humidity_2m,"+
			"surface_pressure,visibility,precipitation",
	)

	requestURL := forecastURL + "?" + params.Encode()
	resp, err := http.Get(requestURL)
	if err != nil {
		return domain.Today{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return domain.Today{}, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var data forecastResponse
	if err = json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return domain.Today{}, err
	}

	updatedAt, err := time.Parse("2006-01-02T15:04", data.Current.Time)
	if err != nil {
		return domain.Today{}, fmt.Errorf("parse current weather time: %w", err)
	}

	return domain.Today{
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
