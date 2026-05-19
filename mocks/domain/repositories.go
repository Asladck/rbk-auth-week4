package mockdomain

import (
	"context"

	"github.com/stretchr/testify/mock"

	"weather-auth-api/internal/domain"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, u *domain.User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	args := m.Called(ctx, id)
	if u := args.Get(0); u != nil {
		return u.(*domain.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if u := args.Get(0); u != nil {
		return u.(*domain.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserRepository) List(ctx context.Context) ([]*domain.User, error) {
	args := m.Called(ctx)
	if v := args.Get(0); v != nil {
		return v.([]*domain.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, u *domain.User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}

func (m *MockUserRepository) SoftDelete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockCityRepository struct {
	mock.Mock
}

func (m *MockCityRepository) Add(ctx context.Context, c *domain.City) error {
	args := m.Called(ctx, c)
	return args.Error(0)
}

func (m *MockCityRepository) ListByUser(ctx context.Context, userID string) ([]*domain.City, error) {
	args := m.Called(ctx, userID)
	if v := args.Get(0); v != nil {
		return v.([]*domain.City), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockCityRepository) Delete(ctx context.Context, cityID, userID string) error {
	args := m.Called(ctx, cityID, userID)
	return args.Error(0)
}

type MockWeatherHistoryRepository struct {
	mock.Mock
}

func (m *MockWeatherHistoryRepository) Save(ctx context.Context, h *domain.WeatherHistory) error {
	args := m.Called(ctx, h)
	return args.Error(0)
}

func (m *MockWeatherHistoryRepository) List(ctx context.Context, userID string, filter domain.HistoryFilter) ([]*domain.WeatherHistory, error) {
	args := m.Called(ctx, userID, filter)
	if v := args.Get(0); v != nil {
		return v.([]*domain.WeatherHistory), args.Error(1)
	}
	return nil, args.Error(1)
}

type MockWeatherClient struct {
	mock.Mock
}

func (m *MockWeatherClient) GetCurrent(ctx context.Context, city string) (*domain.CurrentWeather, error) {
	args := m.Called(ctx, city)
	if v := args.Get(0); v != nil {
		return v.(*domain.CurrentWeather), args.Error(1)
	}
	return nil, args.Error(1)
}

type MockLocationService struct {
	mock.Mock
}

func (m *MockLocationService) GetStates(ctx context.Context, countryCode string) ([]domain.State, error) {
	args := m.Called(ctx, countryCode)
	if v := args.Get(0); v != nil {
		return v.([]domain.State), args.Error(1)
	}
	return nil, args.Error(1)
}

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Register(ctx context.Context, in domain.RegisterInput) (*domain.RegisterResponse, error) {
	args := m.Called(ctx, in)
	if v := args.Get(0); v != nil {
		return v.(*domain.RegisterResponse), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockAuthService) Login(ctx context.Context, in domain.LoginInput) (*domain.TokenResponse, error) {
	args := m.Called(ctx, in)
	if v := args.Get(0); v != nil {
		return v.(*domain.TokenResponse), args.Error(1)
	}
	return nil, args.Error(1)
}
