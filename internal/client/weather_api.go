package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"weather-auth-api/internal/domain"
	"weather-auth-api/pkg/apperrors"
)

const weatherAPIBase = "https://api.weatherapi.com/v1/current.json"

type weatherAPIResponse struct {
	Location struct {
		Name string `json:"name"`
	} `json:"location"`
	Current struct {
		TempC     float64 `json:"temp_c"`
		FeelsLike float64 `json:"feelslike_c"`
		Humidity  int     `json:"humidity"`
		WindKph   float64 `json:"wind_kph"`
		Condition struct {
			Text string `json:"text"`
			Icon string `json:"icon"`
		} `json:"condition"`
	} `json:"current"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type WeatherAPIClient struct {
	apiKey string
	http   *http.Client
}

func NewWeatherAPIClient(apiKey string) *WeatherAPIClient {
	return &WeatherAPIClient{
		apiKey: apiKey,
		http:   &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *WeatherAPIClient) GetCurrent(ctx context.Context, city string) (*domain.CurrentWeather, error) {
	endpoint := fmt.Sprintf("%s?key=%s&q=%s&aqi=no",
		weatherAPIBase,
		c.apiKey,
		url.QueryEscape(city),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("WeatherAPIClient.GetCurrent build request: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.ErrExternalAPI, "weather API unreachable: %v", err)
	}
	defer resp.Body.Close()

	var apiResp weatherAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("WeatherAPIClient.GetCurrent decode: %w", err)
	}

	if resp.StatusCode != http.StatusOK || apiResp.Error != nil {
		msg := "unknown error"
		if apiResp.Error != nil {
			msg = apiResp.Error.Message
		}
		return nil, apperrors.Wrap(apperrors.ErrExternalAPI, "weather API: %s", msg)
	}

	return &domain.CurrentWeather{
		City:      apiResp.Location.Name,
		TempC:     apiResp.Current.TempC,
		FeelsLike: apiResp.Current.FeelsLike,
		Humidity:  apiResp.Current.Humidity,
		WindKph:   apiResp.Current.WindKph,
		Condition: domain.WeatherCondition{
			Text: apiResp.Current.Condition.Text,
			Icon: apiResp.Current.Condition.Icon,
		},
	}, nil
}
