package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"weather-auth-api/internal/domain"
	"weather-auth-api/pkg/apperrors"
)

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, u *domain.User) error {
	u.ID = uuid.NewString()
	now := time.Now().UTC()
	u.CreatedAt = now
	u.UpdatedAt = now

	q := `
			INSERT INTO users (id, name, email, password_hash, role, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.db.Exec(ctx, q,
		u.ID, u.Name, u.Email, u.PasswordHash, u.Role, u.CreatedAt, u.UpdatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return apperrors.Wrap(apperrors.ErrAlreadyExists, "email %s already registered", u.Email)
		}
		return fmt.Errorf("UserRepo.Create: %w", err)
	}
	return nil
}

func (r *UserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	q := `
			SELECT id, name, email, password_hash, role, created_at, updated_at, deleted_at
			FROM users WHERE id = $1`

	return r.scanOne(r.db.QueryRow(ctx, q, id))
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	q := `
			SELECT id, name, email, password_hash, role, created_at, updated_at, deleted_at
			FROM users WHERE email = $1 AND deleted_at IS NULL`

	return r.scanOne(r.db.QueryRow(ctx, q, email))
}

func (r *UserRepo) List(ctx context.Context) ([]*domain.User, error) {
	q := `
			SELECT id, name, email, password_hash, role, created_at, updated_at, deleted_at
			FROM users WHERE deleted_at IS NULL ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("UserRepo.List: %w", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		u := &domain.User{}
		if err := rows.Scan(
			&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role,
			&u.CreatedAt, &u.UpdatedAt, &u.DeletedAt,
		); err != nil {
			return nil, fmt.Errorf("UserRepo.List scan: %w", err)
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *UserRepo) Update(ctx context.Context, u *domain.User) error {
	u.UpdatedAt = time.Now().UTC()
	q := `
			UPDATE users SET name=$1, email=$2, updated_at=$3
			WHERE id=$4 AND deleted_at IS NULL`

	ct, err := r.db.Exec(ctx, q, u.Name, u.Email, u.UpdatedAt, u.ID)
	if err != nil {
		if isUniqueViolation(err) {
			return apperrors.Wrap(apperrors.ErrAlreadyExists, "email %s already registered", u.Email)
		}
		return fmt.Errorf("UserRepo.Update: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return apperrors.Wrap(apperrors.ErrNotFound, "user %s", u.ID)
	}
	return nil
}

func (r *UserRepo) SoftDelete(ctx context.Context, id string) error {
	ct, err := r.db.Exec(ctx,
		`UPDATE users SET deleted_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, id,
	)
	if err != nil {
		return fmt.Errorf("UserRepo.SoftDelete: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return apperrors.Wrap(apperrors.ErrNotFound, "user %s", id)
	}
	return nil
}

func (r *UserRepo) scanOne(row pgx.Row) (*domain.User, error) {
	u := &domain.User{}
	err := row.Scan(
		&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role,
		&u.CreatedAt, &u.UpdatedAt, &u.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("UserRepo.scan: %w", err)
	}
	return u, nil
}
