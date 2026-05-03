package service

import (
	"context"
	"fmt"

	"weather-auth-api/internal/domain"
	"weather-auth-api/pkg/apperrors"
)

type UserSvc struct {
	repo domain.UserRepository
}

func NewUserService(repo domain.UserRepository) *UserSvc {
	return &UserSvc{repo: repo}
}

func (s *UserSvc) Me(ctx context.Context, userID string) (*domain.User, error) {
	u, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("UserSvc.Me: %w", err)
	}
	if u.DeletedAt != nil {
		return nil, apperrors.ErrUserDeleted
	}
	return u, nil
}

func (s *UserSvc) GetByID(ctx context.Context, id string) (*domain.User, error) {
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("UserSvc.GetByID: %w", err)
	}
	return u, nil
}

func (s *UserSvc) List(ctx context.Context) ([]*domain.User, error) {
	users, err := s.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("UserSvc.List: %w", err)
	}
	if users == nil {
		users = []*domain.User{}
	}
	return users, nil
}

func (s *UserSvc) Update(ctx context.Context, id string, in domain.UpdateUserInput) (*domain.User, error) {
	if err := validateUpdateUser(in); err != nil {
		return nil, err
	}

	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("UserSvc.Update: %w", err)
	}
	if u.DeletedAt != nil {
		return nil, apperrors.ErrUserDeleted
	}

	u.Name = in.Name
	u.Email = in.Email

	if err := s.repo.Update(ctx, u); err != nil {
		return nil, fmt.Errorf("UserSvc.Update: %w", err)
	}
	return u, nil
}

func (s *UserSvc) Delete(ctx context.Context, id string) error {
	if err := s.repo.SoftDelete(ctx, id); err != nil {
		return fmt.Errorf("UserSvc.Delete: %w", err)
	}
	return nil
}

func validateUpdateUser(in domain.UpdateUserInput) error {
	if in.Name == "" {
		return &apperrors.ValidationError{Field: "name", Message: "required"}
	}
	if len(in.Name) > 100 {
		return &apperrors.ValidationError{Field: "name", Message: "max 100 characters"}
	}
	if in.Email == "" {
		return &apperrors.ValidationError{Field: "email", Message: "required"}
	}
	if !isValidEmail(in.Email) {
		return &apperrors.ValidationError{Field: "email", Message: "invalid format"}
	}
	return nil
}

func isValidEmail(e string) bool {
	atCount := 0
	for _, ch := range e {
		if ch == '@' {
			atCount++
		}
	}
	return atCount == 1 && len(e) >= 3
}
