package ui

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
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

	elapsed := time.Since(fetchedAt)

	switch {
	case fetchedAt.IsZero():
		str.WriteString(" • время обновления неизвестно")

	case elapsed < time.Minute:
		str.WriteString(" • обновлено только что")

	default:
		minutes := int(elapsed.Minutes())
		str.WriteString(
			fmt.Sprintf(" • обновлено %d мин назад", minutes),
		)
	}

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

func center(value string, width int) string {
	valueWidth := utf8.RuneCountInString(value)

	if valueWidth >= width {
		return value
	}

	padding := width - valueWidth
	left := padding / 2
	right := padding - left

	return strings.Repeat(" ", left) +
		value +
		strings.Repeat(" ", right)
}

func renderDivider(widths []int, indent string) string {
	parts := make([]string, len(widths))

	for i, width := range widths {
		parts[i] = strings.Repeat("-", width)
	}

	return indent + strings.Join(parts, "-+-")
}

func RenderHourly(list []domain.HourlyEntry) string {
	const (
		indent    = "  "
		timeWidth = 6
		tempWidth = 6
		popWidth  = 7
		windWidth = 10
	)

	var builder strings.Builder

	header := fmt.Sprintf(
		"%s%-*s | %*s | %*s | %*s",
		indent,
		timeWidth, "Время",
		tempWidth, "t°C",
		popWidth, "Осадки",
		windWidth, "Ветер м/с",
	)

	divider := indent + strings.Join([]string{
		strings.Repeat("-", timeWidth),
		strings.Repeat("-", tempWidth),
		strings.Repeat("-", popWidth),
		strings.Repeat("-", windWidth),
	}, "-+-")

	builder.WriteString(header)
	builder.WriteByte('\n')
	builder.WriteString(divider)

	for _, item := range list {
		timeValue := fmt.Sprintf(
			"%-*s",
			timeWidth,
			item.Time.Format("15:04"),
		)

		tempValue := colorTemp(
			item.TemperatureC,
			fmt.Sprintf("%*.1f", tempWidth, item.TemperatureC),
		)

		popValue := fmt.Sprintf(
			"%*d%%",
			popWidth-1,
			item.POPPercent,
		)

		windValue := fmt.Sprintf(
			"%*.1f",
			windWidth,
			item.WindSpeedMS,
		)

		builder.WriteByte('\n')
		builder.WriteString(indent)
		builder.WriteString(strings.Join([]string{
			timeValue,
			tempValue,
			popValue,
			windValue,
		}, " | "))
	}

	return builder.String()
}

func RenderDaily(list []domain.DailyEntry) string {
	const indent = "  "

	headers := []string{
		"Дата",
		"Погода",
		"Мин°C",
		"Макс°C",
		"Осадки",
	}

	rows := make([][]string, 0, len(list))

	for _, item := range list {
		dateAndIcon := fmt.Sprintf(
			"%-s %s",
			item.Date.Format("02 Jan"),
			iconForCondition(item.Condition),
		)

		rows = append(rows, []string{
			dateAndIcon,
			item.Condition,

			colorTemp(item.TempMinC, fmt.Sprintf("%.1f", item.TempMinC)),
			colorTemp(item.TempMaxC, fmt.Sprintf("%.1f", item.TempMaxC)),
			fmt.Sprintf("%d%%", item.POPPercent),
		})
	}

	widths := make([]int, len(headers))

	for column, header := range headers {
		widths[column] = utf8.RuneCountInString(header)
	}

	for _, row := range rows {
		for column, value := range row {
			valueWidth := utf8.RuneCountInString(value)

			if valueWidth > widths[column] {
				widths[column] = valueWidth
			}
		}
	}

	renderRow := func(values []string) string {
		cells := make([]string, len(values))

		for column, value := range values {
			switch column {
			case 0, 1:
				cells[column] = fmt.Sprintf(
					"%-*s",
					widths[column],
					center(value, widths[column]),
				)

			default:
				cells[column] = fmt.Sprintf(
					"%*s",
					widths[column],
					center(value, widths[column]),
				)
			}
		}

		return indent + strings.Join(cells, " | ")
	}

	lines := make([]string, 0, len(rows)+2)
	lines = append(lines, renderRow(headers))
	lines = append(lines, renderDivider(widths, indent))

	for _, row := range rows {
		lines = append(lines, renderRow(row))
	}

	return strings.Join(lines, "\n")
}

func RenderMenu() string {
	return fmt.Sprintf("%s────────────────────────────────────────────────────────────%s\n%s[1] Почасовой (12 ч)  [2] На 7 дней  [C] Сменить город  [R] Обновить  [Q] Выход%s",
		cyan, reset, bold, reset)
}
