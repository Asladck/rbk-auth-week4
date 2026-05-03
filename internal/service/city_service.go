package service

import (
	"context"
	"fmt"

	"weather-auth-api/internal/domain"
	"weather-auth-api/pkg/apperrors"
)

type CitySvc struct {
	repo     domain.CityRepository
	userRepo domain.UserRepository
}

func NewCityService(repo domain.CityRepository, userRepo domain.UserRepository) *CitySvc {
	return &CitySvc{repo: repo, userRepo: userRepo}
}

func (s *CitySvc) Add(ctx context.Context, userID string, in domain.AddCityInput) (*domain.City, error) {
	if in.Name == "" {
		return nil, &apperrors.ValidationError{Field: "name", Message: "required"}
	}
	if len(in.Name) > 100 {
		return nil, &apperrors.ValidationError{Field: "name", Message: "max 100 characters"}
	}
	if err := s.assertUserActive(ctx, userID); err != nil {
		return nil, err
	}

	c := &domain.City{UserID: userID, Name: in.Name}
	if err := s.repo.Add(ctx, c); err != nil {
		return nil, fmt.Errorf("CitySvc.Add: %w", err)
	}
	return c, nil
}

func (s *CitySvc) List(ctx context.Context, userID string) ([]*domain.City, error) {
	if err := s.assertUserActive(ctx, userID); err != nil {
		return nil, err
	}
	cities, err := s.repo.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("CitySvc.List: %w", err)
	}
	if cities == nil {
		cities = []*domain.City{}
	}
	return cities, nil
}

func (s *CitySvc) Delete(ctx context.Context, userID, cityID string) error {
	if err := s.assertUserActive(ctx, userID); err != nil {
		return err
	}
	if err := s.repo.Delete(ctx, cityID, userID); err != nil {
		return fmt.Errorf("CitySvc.Delete: %w", err)
	}
	return nil
}

func (s *CitySvc) assertUserActive(ctx context.Context, userID string) error {
	u, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if u.DeletedAt != nil {
		return apperrors.ErrUserDeleted
	}
	return nil
}
