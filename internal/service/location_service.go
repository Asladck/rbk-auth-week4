package service

import (
	"context"
	"fmt"
	"strings"

	"weather-auth-api/internal/domain"
	"weather-auth-api/pkg/apperrors"
)

type LocationSvc struct {
	client domain.LocationClient
}

func NewLocationService(client domain.LocationClient) *LocationSvc {
	return &LocationSvc{client: client}
}

func (s *LocationSvc) GetStates(ctx context.Context, countryCode string) ([]domain.State, error) {
	if countryCode == "" {
		return nil, &apperrors.ValidationError{Field: "country", Message: "required"}
	}

	countryCode = strings.ToUpper(countryCode)

	states, err := s.client.GetStates(ctx, countryCode)
	if err != nil {
		return nil, fmt.Errorf("LocationSvc.GetStates: %w", err)
	}
	return states, nil
}
