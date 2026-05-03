package domain

import "time"

const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

type User struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"`
	Role         string     `json:"role"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}

type RegisterInput struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

type RegisterResponse struct {
	Message string `json:"message"`
}

type UpdateUserInput struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type City struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type AddCityInput struct {
	Name string `json:"name"`
}

type WeatherCondition struct {
	Text string `json:"text"`
	Icon string `json:"icon"`
}

type CurrentWeather struct {
	City      string           `json:"city"`
	TempC     float64          `json:"temp_c"`
	FeelsLike float64          `json:"feels_like_c"`
	Humidity  int              `json:"humidity"`
	WindKph   float64          `json:"wind_kph"`
	Condition WeatherCondition `json:"condition"`
}

type WeatherResponse struct {
	UserID  string           `json:"user_id"`
	Results []CurrentWeather `json:"results"`
}

type WeatherHistory struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	City        string    `json:"city"`
	TempC       float64   `json:"temp_c"`
	FeelsLike   float64   `json:"feels_like_c"`
	Humidity    int       `json:"humidity"`
	WindKph     float64   `json:"wind_kph"`
	Condition   string    `json:"condition"`
	RequestedAt time.Time `json:"requested_at"`
}

type HistoryFilter struct {
	City  string
	Limit int
}

type State struct {
	Name      string `json:"name"`
	StateCode string `json:"iso2"`
}
