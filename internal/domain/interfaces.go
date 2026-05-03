package domain

import "context"

type UserRepository interface {
	Create(ctx context.Context, u *User) error
	GetByID(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	List(ctx context.Context) ([]*User, error)
	Update(ctx context.Context, u *User) error
	SoftDelete(ctx context.Context, id string) error
}

type CityRepository interface {
	Add(ctx context.Context, c *City) error
	ListByUser(ctx context.Context, userID string) ([]*City, error)
	Delete(ctx context.Context, cityID, userID string) error
}

type WeatherHistoryRepository interface {
	Save(ctx context.Context, h *WeatherHistory) error
	List(ctx context.Context, userID string, filter HistoryFilter) ([]*WeatherHistory, error)
}

type WeatherClient interface {
	GetCurrent(ctx context.Context, city string) (*CurrentWeather, error)
}

type LocationClient interface {
	GetStates(ctx context.Context, countryCode string) ([]State, error)
}

type AuthService interface {
	Register(ctx context.Context, in RegisterInput) (*RegisterResponse, error)
	Login(ctx context.Context, in LoginInput) (*TokenResponse, error)
}

type UserService interface {
	Me(ctx context.Context, userID string) (*User, error)
	GetByID(ctx context.Context, id string) (*User, error)
	List(ctx context.Context) ([]*User, error)
	Update(ctx context.Context, id string, in UpdateUserInput) (*User, error)
	Delete(ctx context.Context, id string) error
}

type CityService interface {
	Add(ctx context.Context, userID string, in AddCityInput) (*City, error)
	List(ctx context.Context, userID string) ([]*City, error)
	Delete(ctx context.Context, userID, cityID string) error
}

type WeatherService interface {
	GetWeatherForUser(ctx context.Context, userID string) (*WeatherResponse, error)
	GetHistory(ctx context.Context, userID string, filter HistoryFilter) ([]*WeatherHistory, error)
}

type LocationService interface {
	GetStates(ctx context.Context, countryCode string) ([]State, error)
}
