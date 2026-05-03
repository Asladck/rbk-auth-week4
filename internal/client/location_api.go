package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"weather-auth-api/internal/domain"
	"weather-auth-api/pkg/apperrors"
)

const locationAPIBase = "https://api.countrystatecity.in/v1"

type locationStateResponse struct {
	Name string `json:"name"`
	Iso2 string `json:"iso2"`
}

type LocationAPIClient struct {
	apiKey string
	http   *http.Client
}

func NewLocationAPIClient(apiKey string) *LocationAPIClient {
	return &LocationAPIClient{
		apiKey: apiKey,
		http:   &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *LocationAPIClient) GetStates(ctx context.Context, countryCode string) ([]domain.State, error) {
	endpoint := fmt.Sprintf("%s/countries/%s/states", locationAPIBase, countryCode)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("LocationAPIClient.GetStates build request: %w", err)
	}
	req.Header.Set("X-CSCAPI-KEY", c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.ErrExternalAPI, "location API unreachable: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, apperrors.Wrap(apperrors.ErrNotFound, "country %s not found", countryCode)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, apperrors.Wrap(apperrors.ErrExternalAPI, "location API returned %d", resp.StatusCode)
	}

	var raw []locationStateResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("LocationAPIClient.GetStates decode: %w", err)
	}

	states := make([]domain.State, 0, len(raw))
	for _, s := range raw {
		states = append(states, domain.State{Name: s.Name, StateCode: s.Iso2})
	}
	return states, nil
}
