package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"weather-auth-api/internal/domain"
	"weather-auth-api/pkg/apperrors"
)

type CityRepo struct {
	db *pgxpool.Pool
}

func NewCityRepo(db *pgxpool.Pool) *CityRepo {
	return &CityRepo{db: db}
}

func (r *CityRepo) Add(ctx context.Context, c *domain.City) error {
	c.ID = uuid.NewString()
	c.CreatedAt = time.Now().UTC()

	q := `INSERT INTO cities (id, user_id, name, created_at) VALUES ($1, $2, $3, $4)`
	_, err := r.db.Exec(ctx, q, c.ID, c.UserID, c.Name, c.CreatedAt)
	if err != nil {
		if isFKViolation(err) {
			return apperrors.Wrap(apperrors.ErrNotFound, "user %s", c.UserID)
		}
		if isUniqueViolation(err) {
			return apperrors.Wrap(apperrors.ErrAlreadyExists, "city %s already added", c.Name)
		}
		return fmt.Errorf("CityRepo.Add: %w", err)
	}
	return nil
}

func (r *CityRepo) ListByUser(ctx context.Context, userID string) ([]*domain.City, error) {
	q := `
		SELECT id, user_id, name, created_at FROM cities
		WHERE user_id = $1 ORDER BY created_at ASC`

	rows, err := r.db.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("CityRepo.ListByUser: %w", err)
	}
	defer rows.Close()

	var cities []*domain.City
	for rows.Next() {
		c := &domain.City{}
		if err := rows.Scan(&c.ID, &c.UserID, &c.Name, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("CityRepo.ListByUser scan: %w", err)
		}
		cities = append(cities, c)
	}
	return cities, rows.Err()
}

func (r *CityRepo) Delete(ctx context.Context, cityID, userID string) error {
	ct, err := r.db.Exec(ctx,
		`DELETE FROM cities WHERE id=$1 AND user_id=$2`, cityID, userID,
	)
	if err != nil {
		return fmt.Errorf("CityRepo.Delete: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return apperrors.Wrap(apperrors.ErrNotFound, "city %s for user %s", cityID, userID)
	}
	return nil
}
