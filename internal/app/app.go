package app

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
	"weather-cli/internal/cache"
	"weather-cli/internal/config"
	"weather-cli/internal/domain"
	"weather-cli/internal/provider"
	"weather-cli/internal/provider/openmeteo"
	"weather-cli/internal/ui"
)

const (
	cityValidationTimeout = 60 * time.Second
	todayCacheTTL         = 10 * time.Minute
	hourlyForecastHours   = 12
	hourlyCacheTTL        = 30 * time.Minute
	dailyForecastDays     = 7
	dailyCacheTTL         = 60 * time.Minute
	minRefreshInterval    = 10 * time.Second
)

type screen int

const (
	todayScreen screen = iota
	hourlyScreen
	dailyScreen
	changeCityScreen
)

type application struct {
	config        config.Config
	cache         *cache.TTLCache
	provider      provider.WeatherProvider
	city          string
	displayCity   string
	reader        *bufio.Reader
	lastRefreshAt time.Time
}

func Run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	client := openmeteo.NewClient()

	ttlCache := cache.New()

	app := &application{
		config:      cfg,
		cache:       ttlCache,
		provider:    client,
		city:        cfg.DefaultCity,
		displayCity: strings.TrimSpace(cfg.DefaultCity),
		reader:      bufio.NewReader(os.Stdin),
	}

	ctx := context.Background()

	if app.city == "" {
		if err = app.promptCity(ctx, false); err != nil {
			return err
		}
	}

	fmt.Println("Город по умолчанию:", app.city)

	return app.loop(ctx)
}

func (a *application) promptCity(ctx context.Context, isCityChangeMode bool) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		if isCityChangeMode {
			fmt.Println()
			fmt.Println("Введите новый город: ")
		} else {
			fmt.Println()
			fmt.Println("Введите город по умолчанию: ")
		}

		city, err := a.reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("error read city: %w", err)
		}
		city = strings.TrimSpace(city)

		if city == "" {
			fmt.Println("Название города не должно быть пустым.")
			continue
		}

		requestCtx, cancel := context.WithTimeout(ctx, cityValidationTimeout)

		today, err := a.provider.GetToday(requestCtx, city)
		cancel()

		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}

			fmt.Printf("Не удалось проверить город: %v\n", err)
			fmt.Println("Попробуйте ещё раз.")
			continue
		}

		nextConfig := a.config
		nextConfig.DefaultCity = city

		if err = config.Save(nextConfig); err != nil {
			return fmt.Errorf("save default city: %w", err)
		}

		a.config = nextConfig
		a.city = city
		a.displayCity = today.City

		a.cache.Set("today:"+city, today, todayCacheTTL)

		return nil
	}
}

func (a *application) loadToday(ctx context.Context, force bool) (domain.Today, bool, time.Time, error) {
	if err := ctx.Err(); err != nil {
		return domain.Today{}, false, time.Time{}, err
	}

	key := "today:" + a.city
	if !force {
		value, fetchedAt, found := a.cache.Get(key)
		if found {
			today, typeOK := value.(domain.Today)
			if typeOK {
				return today, true, fetchedAt, nil
			}
		}
	}

	requestCtx, cancel := context.WithTimeout(ctx, cityValidationTimeout)
	defer cancel()

	today, err := a.provider.GetToday(requestCtx, a.city)
	if err != nil {
		return domain.Today{}, false, time.Time{}, fmt.Errorf("get today weather: %w", err)
	}

	fetchedAt := a.cache.Set(key, today, todayCacheTTL)

	return today, false, fetchedAt, nil
}

func (a *application) loadHourly(ctx context.Context, force bool) ([]domain.HourlyEntry, bool, time.Time, error) {
	if err := ctx.Err(); err != nil {
		return nil, false, time.Time{}, err
	}

	key := fmt.Sprintf("hourly:%s:%d", a.city, hourlyForecastHours)
	if !force {
		value, fetchedAt, found := a.cache.Get(key)
		if found {
			hourly, typeOk := value.([]domain.HourlyEntry)
			if typeOk {
				return hourly, true, fetchedAt, nil
			}
		}
	}

	requestCtx, cancel := context.WithTimeout(ctx, cityValidationTimeout)
	defer cancel()

	hourly, err := a.provider.GetHourly(requestCtx, a.city, hourlyForecastHours)
	if err != nil {
		return nil, false, time.Time{}, fmt.Errorf("get hourly weather: %w", err)
	}

	fetchedAt := a.cache.Set(key, hourly, hourlyCacheTTL)

	return hourly, false, fetchedAt, nil
}

func (a *application) loadDaily(ctx context.Context, force bool) ([]domain.DailyEntry, bool, time.Time, error) {
	if err := ctx.Err(); err != nil {
		return nil, false, time.Time{}, err
	}

	key := fmt.Sprintf("daily:%s:%d", a.city, dailyForecastDays)
	if !force {
		value, fetchedAt, found := a.cache.Get(key)
		if found {
			daily, typeOk := value.([]domain.DailyEntry)
			if typeOk {
				return daily, true, fetchedAt, nil
			}
		}
	}

	requestCtx, cancel := context.WithTimeout(ctx, cityValidationTimeout)
	defer cancel()

	daily, err := a.provider.GetDaily(requestCtx, a.city, dailyForecastDays)
	if err != nil {
		return nil, false, time.Time{}, fmt.Errorf("get daily weather: %w", err)
	}

	fetchedAt := a.cache.Set(key, daily, dailyCacheTTL)

	return daily, false, fetchedAt, nil
}

func (a *application) loop(ctx context.Context) error {
	currentScreen := todayScreen
	forceNextRefresh := false

	for {
		if err := ctx.Err(); err != nil {
			return ctx.Err()
		}

		switch currentScreen {
		case todayScreen:
			today, cached, fetchedAt, err := a.loadToday(ctx, forceNextRefresh)
			if err != nil {
				if ctxErr := ctx.Err(); ctxErr != nil {
					return ctxErr
				}

				fmt.Printf("Не удалось загрузить текущую погоду: %v\n", err)
			} else {
				fmt.Println()
				fmt.Println(ui.Header(a.displayCity, cached, fetchedAt))
				fmt.Println(ui.RenderToday(today))

				if forceNextRefresh {
					a.lastRefreshAt = time.Now()
					forceNextRefresh = false
				}
			}

		case hourlyScreen:
			hourly, cached, fetchedAt, err := a.loadHourly(ctx, forceNextRefresh)
			if err != nil {
				if ctxErr := ctx.Err(); ctxErr != nil {
					return ctxErr
				}

				fmt.Printf("Не удалось загрузить почасовой прогноз: %v\n", err)
			} else {
				fmt.Println()
				fmt.Println(ui.Header(a.displayCity, cached, fetchedAt))
				fmt.Println(ui.RenderHourly(hourly))

				if forceNextRefresh {
					a.lastRefreshAt = time.Now()
					forceNextRefresh = false
				}
			}

		case dailyScreen:
			daily, cached, fetchedAt, err := a.loadDaily(ctx, forceNextRefresh)
			if err != nil {
				if ctxErr := ctx.Err(); ctxErr != nil {
					return ctxErr
				}

				fmt.Printf("Не удалось загрузить почасовой прогноз: %v\n", err)
			} else {
				fmt.Println()
				fmt.Println(ui.Header(a.displayCity, cached, fetchedAt))
				fmt.Println(ui.RenderDaily(daily))

				if forceNextRefresh {
					a.lastRefreshAt = time.Now()
					forceNextRefresh = false
				}
			}

		case changeCityScreen:
			if err := a.promptCity(ctx, true); err != nil {
				fmt.Printf("error change city: %v\n", err)
				continue
			}

			currentScreen = todayScreen
			continue
		}

		fmt.Println(ui.RenderMenu())

		fmt.Print("> ")

		input, readErr := a.reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if errors.Is(readErr, io.EOF) && len(input) == 0 {
			fmt.Println("\nВвод завершён.")
			return nil
		}

		command := strings.ToLower(strings.TrimSpace(input))

		switch command {
		case "0":
			currentScreen = todayScreen
		case "1":
			currentScreen = hourlyScreen
		case "2":
			currentScreen = dailyScreen
		case "c":
			currentScreen = changeCityScreen
		case "r":
			now := time.Now()
			if now.Sub(a.lastRefreshAt) < minRefreshInterval {
				fmt.Printf(
					"%s⚠️ Обновление доступно не чаще чем раз в %v.%s\n",
					"\u001B[33m",
					minRefreshInterval,
					"\u001B[0m",
				)
				continue
			}

			forceNextRefresh = true
			continue
		case "q":
			fmt.Println("Выход...")
			return nil
		case "":

		default:
			fmt.Println("Неизвестная команда")
			continue
		}

		if errors.Is(readErr, io.EOF) {
			fmt.Println("Ввод завершён.")
			return nil
		}
	}
}
