package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"weather-auth-api/internal/domain"
	mockdomain "weather-auth-api/mocks/domain"
	"weather-auth-api/pkg/apperrors"
)

func TestWeatherSvc_GetWeatherForUser(t *testing.T) {
	now := time.Now().UTC()

	tcs := []struct {
		name      string
		userID    string
		setup     func(userRepo *mockdomain.MockUserRepository, cityRepo *mockdomain.MockCityRepository, histRepo *mockdomain.MockWeatherHistoryRepository, client *mockdomain.MockWeatherClient)
		assertion func(t *testing.T, got *domain.WeatherResponse, err error)
	}{
		{
			name:   "invalid id",
			userID: "",
			setup: func(userRepo *mockdomain.MockUserRepository, cityRepo *mockdomain.MockCityRepository, histRepo *mockdomain.MockWeatherHistoryRepository, client *mockdomain.MockWeatherClient) {
				userRepo.On("GetByID", mock.Anything, "").Return((*domain.User)(nil), apperrors.ErrNotFound).Once()
			},
			assertion: func(t *testing.T, got *domain.WeatherResponse, err error) {
				require.Error(t, err)
				assert.Nil(t, got)
			},
		},
		{
			name:   "entity not found",
			userID: "u1",
			setup: func(userRepo *mockdomain.MockUserRepository, cityRepo *mockdomain.MockCityRepository, histRepo *mockdomain.MockWeatherHistoryRepository, client *mockdomain.MockWeatherClient) {
				userRepo.On("GetByID", mock.Anything, "u1").Return((*domain.User)(nil), apperrors.ErrNotFound).Once()
			},
			assertion: func(t *testing.T, got *domain.WeatherResponse, err error) {
				require.Error(t, err)
				assert.Nil(t, got)
			},
		},
		{
			name:   "repository failure",
			userID: "u1",
			setup: func(userRepo *mockdomain.MockUserRepository, cityRepo *mockdomain.MockCityRepository, histRepo *mockdomain.MockWeatherHistoryRepository, client *mockdomain.MockWeatherClient) {
				userRepo.On("GetByID", mock.Anything, "u1").Return((*domain.User)(nil), errors.New("db down")).Once()
			},
			assertion: func(t *testing.T, got *domain.WeatherResponse, err error) {
				require.Error(t, err)
				assert.Nil(t, got)
			},
		},
		{
			name:   "user deleted",
			userID: "u1",
			setup: func(userRepo *mockdomain.MockUserRepository, cityRepo *mockdomain.MockCityRepository, histRepo *mockdomain.MockWeatherHistoryRepository, client *mockdomain.MockWeatherClient) {
				userRepo.On("GetByID", mock.Anything, "u1").Return(&domain.User{ID: "u1", DeletedAt: &now}, nil).Once()
			},
			assertion: func(t *testing.T, got *domain.WeatherResponse, err error) {
				require.Error(t, err)
				assert.ErrorIs(t, err, apperrors.ErrUserDeleted)
				assert.Nil(t, got)
			},
		},
		{
			name:   "empty data",
			userID: "u1",
			setup: func(userRepo *mockdomain.MockUserRepository, cityRepo *mockdomain.MockCityRepository, histRepo *mockdomain.MockWeatherHistoryRepository, client *mockdomain.MockWeatherClient) {
				userRepo.On("GetByID", mock.Anything, "u1").Return(&domain.User{ID: "u1"}, nil).Once()
				cityRepo.On("ListByUser", mock.Anything, "u1").Return([]*domain.City{}, nil).Once()
			},
			assertion: func(t *testing.T, got *domain.WeatherResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, got)
				assert.Equal(t, "u1", got.UserID)
				assert.NotNil(t, got.Results)
				assert.Len(t, got.Results, 0)
			},
		},
		{
			name:   "internal error from cities repository",
			userID: "u1",
			setup: func(userRepo *mockdomain.MockUserRepository, cityRepo *mockdomain.MockCityRepository, histRepo *mockdomain.MockWeatherHistoryRepository, client *mockdomain.MockWeatherClient) {
				userRepo.On("GetByID", mock.Anything, "u1").Return(&domain.User{ID: "u1"}, nil).Once()
				cityRepo.On("ListByUser", mock.Anything, "u1").Return(([]*domain.City)(nil), errors.New("db down")).Once()
			},
			assertion: func(t *testing.T, got *domain.WeatherResponse, err error) {
				require.Error(t, err)
				assert.Nil(t, got)
			},
		},
		{
			name:   "happy path",
			userID: "u1",
			setup: func(userRepo *mockdomain.MockUserRepository, cityRepo *mockdomain.MockCityRepository, histRepo *mockdomain.MockWeatherHistoryRepository, client *mockdomain.MockWeatherClient) {
				userRepo.On("GetByID", mock.Anything, "u1").Return(&domain.User{ID: "u1"}, nil).Once()
				cityRepo.On("ListByUser", mock.Anything, "u1").Return([]*domain.City{{Name: "Almaty"}}, nil).Once()
				client.On("GetCurrent", mock.Anything, "Almaty").Return(&domain.CurrentWeather{City: "Almaty", TempC: 1.2, FeelsLike: 1.0, Humidity: 10, WindKph: 2.2, Condition: domain.WeatherCondition{Text: "Ok"}}, nil).Once()
				histRepo.On("Save", mock.Anything, mock.MatchedBy(func(h *domain.WeatherHistory) bool {
					return h != nil && h.UserID == "u1" && h.City == "Almaty" && h.Condition == "Ok" && !h.RequestedAt.IsZero()
				})).Return(nil).Once()
			},
			assertion: func(t *testing.T, got *domain.WeatherResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, got)
				assert.Len(t, got.Results, 1)
				assert.Equal(t, "Almaty", got.Results[0].City)
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			userRepo := new(mockdomain.MockUserRepository)
			cityRepo := new(mockdomain.MockCityRepository)
			histRepo := new(mockdomain.MockWeatherHistoryRepository)
			client := new(mockdomain.MockWeatherClient)
			log := zaptest.NewLogger(t)

			svc := NewWeatherService(userRepo, cityRepo, histRepo, client, log)

			tc.setup(userRepo, cityRepo, histRepo, client)

			got, err := svc.GetWeatherForUser(context.Background(), tc.userID)

			time.Sleep(30 * time.Millisecond)

			userRepo.AssertExpectations(t)
			cityRepo.AssertExpectations(t)
			histRepo.AssertExpectations(t)
			client.AssertExpectations(t)

			tc.assertion(t, got, err)
		})
	}
}

func TestWeatherSvc_GetHistory(t *testing.T) {
	now := time.Now().UTC()

	tcs := []struct {
		name      string
		userID    string
		filter    domain.HistoryFilter
		setup     func(userRepo *mockdomain.MockUserRepository, histRepo *mockdomain.MockWeatherHistoryRepository)
		assertion func(t *testing.T, got []*domain.WeatherHistory, err error)
	}{
		{
			name:   "invalid input",
			userID: "u1",
			filter: domain.HistoryFilter{City: ""},
			setup:  func(userRepo *mockdomain.MockUserRepository, histRepo *mockdomain.MockWeatherHistoryRepository) {},
			assertion: func(t *testing.T, got []*domain.WeatherHistory, err error) {
				require.Error(t, err)
				assert.ErrorIs(t, err, apperrors.ErrInvalidInput)
				assert.Nil(t, got)
			},
		},
		{
			name:   "invalid id",
			userID: "",
			filter: domain.HistoryFilter{City: "Almaty", Limit: 10},
			setup: func(userRepo *mockdomain.MockUserRepository, histRepo *mockdomain.MockWeatherHistoryRepository) {
				userRepo.On("GetByID", mock.Anything, "").Return((*domain.User)(nil), apperrors.ErrNotFound).Once()
			},
			assertion: func(t *testing.T, got []*domain.WeatherHistory, err error) {
				require.Error(t, err)
				assert.Nil(t, got)
			},
		},
		{
			name:   "entity not found",
			userID: "u1",
			filter: domain.HistoryFilter{City: "Almaty", Limit: 10},
			setup: func(userRepo *mockdomain.MockUserRepository, histRepo *mockdomain.MockWeatherHistoryRepository) {
				userRepo.On("GetByID", mock.Anything, "u1").Return((*domain.User)(nil), apperrors.ErrNotFound).Once()
			},
			assertion: func(t *testing.T, got []*domain.WeatherHistory, err error) {
				require.Error(t, err)
				assert.Nil(t, got)
			},
		},
		{
			name:   "user deleted",
			userID: "u1",
			filter: domain.HistoryFilter{City: "Almaty", Limit: 10},
			setup: func(userRepo *mockdomain.MockUserRepository, histRepo *mockdomain.MockWeatherHistoryRepository) {
				userRepo.On("GetByID", mock.Anything, "u1").Return(&domain.User{ID: "u1", DeletedAt: &now}, nil).Once()
			},
			assertion: func(t *testing.T, got []*domain.WeatherHistory, err error) {
				require.Error(t, err)
				assert.ErrorIs(t, err, apperrors.ErrUserDeleted)
				assert.Nil(t, got)
			},
		},
		{
			name:   "repository failure",
			userID: "u1",
			filter: domain.HistoryFilter{City: "Almaty", Limit: 10},
			setup: func(userRepo *mockdomain.MockUserRepository, histRepo *mockdomain.MockWeatherHistoryRepository) {
				userRepo.On("GetByID", mock.Anything, "u1").Return(&domain.User{ID: "u1"}, nil).Once()
				histRepo.On("List", mock.Anything, "u1", domain.HistoryFilter{City: "Almaty", Limit: 10}).Return(([]*domain.WeatherHistory)(nil), errors.New("db down")).Once()
			},
			assertion: func(t *testing.T, got []*domain.WeatherHistory, err error) {
				require.Error(t, err)
				assert.Nil(t, got)
			},
		},
		{
			name:   "empty data",
			userID: "u1",
			filter: domain.HistoryFilter{City: "Almaty", Limit: 10},
			setup: func(userRepo *mockdomain.MockUserRepository, histRepo *mockdomain.MockWeatherHistoryRepository) {
				userRepo.On("GetByID", mock.Anything, "u1").Return(&domain.User{ID: "u1"}, nil).Once()
				histRepo.On("List", mock.Anything, "u1", domain.HistoryFilter{City: "Almaty", Limit: 10}).Return(([]*domain.WeatherHistory)(nil), nil).Once()
			},
			assertion: func(t *testing.T, got []*domain.WeatherHistory, err error) {
				require.NoError(t, err)
				assert.NotNil(t, got)
				assert.Len(t, got, 0)
			},
		},
		{
			name:   "happy path",
			userID: "u1",
			filter: domain.HistoryFilter{City: "Almaty", Limit: 10},
			setup: func(userRepo *mockdomain.MockUserRepository, histRepo *mockdomain.MockWeatherHistoryRepository) {
				userRepo.On("GetByID", mock.Anything, "u1").Return(&domain.User{ID: "u1"}, nil).Once()
				histRepo.On("List", mock.Anything, "u1", domain.HistoryFilter{City: "Almaty", Limit: 10}).Return([]*domain.WeatherHistory{{ID: "h1"}}, nil).Once()
			},
			assertion: func(t *testing.T, got []*domain.WeatherHistory, err error) {
				require.NoError(t, err)
				assert.Len(t, got, 1)
				assert.Equal(t, "h1", got[0].ID)
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			userRepo := new(mockdomain.MockUserRepository)
			histRepo := new(mockdomain.MockWeatherHistoryRepository)
			log := zaptest.NewLogger(t)

			svc := NewWeatherService(userRepo, new(mockdomain.MockCityRepository), histRepo, new(mockdomain.MockWeatherClient), log)

			tc.setup(userRepo, histRepo)
			got, err := svc.GetHistory(context.Background(), tc.userID, tc.filter)

			userRepo.AssertExpectations(t)
			histRepo.AssertExpectations(t)

			tc.assertion(t, got, err)
		})
	}
}
