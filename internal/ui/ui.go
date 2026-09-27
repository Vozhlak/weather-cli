package ui

import (
	"fmt"
	"strings"
	"time"
	"weather-cli/internal/domain"
)

const (
	reset  = "\x1b[0m"
	bold   = "\x1b[1m"
	red    = "\x1b[31m"
	blue   = "\x1b[34m"
	yellow = "\x1b[33m"
	gray   = "\x1b[38;5;244m"
	cyan   = "\x1b[36m"
)

func colorTemp(celsius float64, text string) string {
	var color string

	switch {
	case celsius < 0:
		color = blue
	case celsius <= 25:
		color = yellow
	default:
		color = red
	}

	return fmt.Sprintf("%s%s%s", color, text, reset)
}

func iconForCondition(cond string) string {
	cond = strings.ToLower(cond)
	switch {
	case strings.Contains(cond, "гроза"):
		return "⛈"
	case strings.Contains(cond, "дождь"):
		return "🌧"
	case strings.Contains(cond, "снег"):
		return "❄"
	case strings.Contains(cond, "ясно"):
		return "☀"
	case strings.Contains(cond, "облач"):
		return "☁"
	default:
		return "🌡"
	}
}

func Header(city string, cached bool, fetchedAt time.Time) string {
	var str strings.Builder

	str.WriteString(fmt.Sprintf("%s%s%s", bold, city, reset))

	now := time.Now()
	diffMinutes := int(now.Sub(fetchedAt).Seconds() / 60)

	str.WriteString(fmt.Sprintf(" • обновлено %d мин назад", diffMinutes))

	if cached {
		str.WriteString(fmt.Sprintf(" • %sиз кэша%s", gray, reset))
	}

	return str.String()
}

func RenderToday(t domain.Today) string {
	var str strings.Builder

	str.WriteString(fmt.Sprintf("Сегодня в %s [%s]\n", t.City, iconForCondition(t.Condition)))
	str.WriteString(fmt.Sprintf("  Температура:      %s°C (ощущается как %.1f°C)\n",
		colorTemp(t.TemperatureC, fmt.Sprintf("%.1f", t.TemperatureC)), t.FeelsLikeC))
	str.WriteString(fmt.Sprintf("  Условие:          %s\n", t.Condition))
	str.WriteString(fmt.Sprintf("  Ветер:            %.1f м/с (%d°)\n", t.WindSpeedMS, t.WindDirectionDeg))
	str.WriteString(fmt.Sprintf("  Влажность:        %d%%\n", t.HumidityPercent))
	str.WriteString(fmt.Sprintf("  Давление:         %d hPa\n", t.PressureHPa))
	str.WriteString(fmt.Sprintf("  Видимость:        %.1f км\n", t.VisibilityKm))
	str.WriteString(fmt.Sprintf("  Осадки (1ч):      %.1f мм\n", t.PrecipitationMm))

	return str.String()
}

func RenderMenu() string {
	return fmt.Sprintf("%s────────────────────────────────────────────────────────────%s\n%s[1] Почасовой (12 ч)  [2] На 7 дней  [C] Сменить город  [R] Обновить  [Q] Выход%s",
		cyan, reset, bold, reset)
}
