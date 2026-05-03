package service

import (
	"context"
	"fmt"
	"sync"

	"go.uber.org/zap"

	"weather-auth-api/internal/domain"
	"weather-auth-api/pkg/apperrors"
)

type WeatherSvc struct {
	userRepo    domain.UserRepository
	cityRepo    domain.CityRepository
	historyRepo domain.WeatherHistoryRepository
	client      domain.WeatherClient
	log         *zap.Logger
}

func NewWeatherService(
	userRepo domain.UserRepository,
	cityRepo domain.CityRepository,
	historyRepo domain.WeatherHistoryRepository,
	client domain.WeatherClient,
	log *zap.Logger,
) *WeatherSvc {
	return &WeatherSvc{
		userRepo:    userRepo,
		cityRepo:    cityRepo,
		historyRepo: historyRepo,
		client:      client,
		log:         log,
	}
}

func (s *WeatherSvc) GetWeatherForUser(ctx context.Context, userID string) (*domain.WeatherResponse, error) {
	u, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("WeatherSvc.GetWeatherForUser: %w", err)
	}
	if u.DeletedAt != nil {
		return nil, apperrors.ErrUserDeleted
	}

	cities, err := s.cityRepo.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("WeatherSvc: list cities: %w", err)
	}
	if len(cities) == 0 {
		return &domain.WeatherResponse{UserID: userID, Results: []domain.CurrentWeather{}}, nil
	}

	type result struct {
		weather *domain.CurrentWeather
		err     error
		city    string
	}

	results := make([]result, len(cities))
	var wg sync.WaitGroup

	for i, c := range cities {
		wg.Add(1)
		go func(idx int, cityName string) {
			defer wg.Done()
			w, e := s.client.GetCurrent(ctx, cityName)
			results[idx] = result{weather: w, err: e, city: cityName}
		}(i, c.Name)
	}
	wg.Wait()

	var weatherResults []domain.CurrentWeather
	for _, r := range results {
		if r.err != nil {
			s.log.Warn("weather fetch failed for city",
				zap.String("city", r.city),
				zap.Error(r.err),
			)
			continue
		}
		weatherResults = append(weatherResults, *r.weather)

		// Persist history best-effort — use background ctx so request cancellation
		// doesn't abort the DB write.
		go func(w domain.CurrentWeather) {
			h := &domain.WeatherHistory{
				UserID:    userID,
				City:      w.City,
				TempC:     w.TempC,
				FeelsLike: w.FeelsLike,
				Humidity:  w.Humidity,
				WindKph:   w.WindKph,
				Condition: w.Condition.Text,
			}
			if err := s.historyRepo.Save(context.Background(), h); err != nil {
				s.log.Error("failed to save weather history",
					zap.String("city", w.City),
					zap.Error(err),
				)
			}
		}(*r.weather)
	}

	if weatherResults == nil {
		weatherResults = []domain.CurrentWeather{}
	}
	return &domain.WeatherResponse{UserID: userID, Results: weatherResults}, nil
}

func (s *WeatherSvc) GetHistory(ctx context.Context, userID string, filter domain.HistoryFilter) ([]*domain.WeatherHistory, error) {
	if filter.City == "" {
		return nil, &apperrors.ValidationError{Field: "city", Message: "required query parameter"}
	}
	u, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("WeatherSvc.GetHistory: %w", err)
	}
	if u.DeletedAt != nil {
		return nil, apperrors.ErrUserDeleted
	}

	history, err := s.historyRepo.List(ctx, userID, filter)
	if err != nil {
		return nil, fmt.Errorf("WeatherSvc.GetHistory: %w", err)
	}
	if history == nil {
		history = []*domain.WeatherHistory{}
	}
	return history, nil
}
