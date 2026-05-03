package service

import (
	"context"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"weather-auth-api/internal/domain"
	"weather-auth-api/pkg/apperrors"
	"weather-auth-api/pkg/jwtutil"
)

type AuthSvc struct {
	userRepo domain.UserRepository
	jwt      *jwtutil.Manager
}

func NewAuthService(userRepo domain.UserRepository, jwt *jwtutil.Manager) *AuthSvc {
	return &AuthSvc{userRepo: userRepo, jwt: jwt}
}

func (s *AuthSvc) Register(ctx context.Context, in domain.RegisterInput) (*domain.RegisterResponse, error) {
	if err := validateRegister(in); err != nil {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), 12)
	if err != nil {
		return nil, fmt.Errorf("AuthSvc.Register bcrypt: %w", err)
	}

	u := &domain.User{
		Name:         in.Name,
		Email:        in.Email,
		PasswordHash: string(hash),
		Role:         domain.RoleUser,
	}

	if err := s.userRepo.Create(ctx, u); err != nil {
		return nil, fmt.Errorf("AuthSvc.Register: %w", err)
	}

	return &domain.RegisterResponse{Message: "user created successfully"}, nil
}

func (s *AuthSvc) Login(ctx context.Context, in domain.LoginInput) (*domain.TokenResponse, error) {
	if in.Email == "" || in.Password == "" {
		return nil, &apperrors.ValidationError{Field: "credentials", Message: "email and password required"}
	}

	u, err := s.userRepo.GetByEmail(ctx, in.Email)
	if err != nil {
		return nil, apperrors.ErrBadCreds
	}

	if u.DeletedAt != nil {
		return nil, apperrors.ErrBadCreds
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(in.Password)); err != nil {
		return nil, apperrors.ErrBadCreds
	}

	return s.issueToken(u)
}

func (s *AuthSvc) issueToken(u *domain.User) (*domain.TokenResponse, error) {
	token, err := s.jwt.Generate(u.ID, u.Email, u.Role)
	if err != nil {
		return nil, fmt.Errorf("AuthSvc.issueToken: %w", err)
	}
	return &domain.TokenResponse{AccessToken: token, TokenType: "Bearer"}, nil
}

func validateRegister(in domain.RegisterInput) error {
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
	if len(in.Password) < 8 {
		return &apperrors.ValidationError{Field: "password", Message: "minimum 8 characters"}
	}
	return nil
}
