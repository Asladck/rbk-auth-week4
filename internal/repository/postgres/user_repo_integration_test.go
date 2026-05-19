package postgres

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"weather-auth-api/internal/domain"
	"weather-auth-api/pkg/apperrors"
)

func TestUserRepo_Integration_CreateAndGet(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)
	defer pool.Close()

	require.NoError(t, pool.Ping(ctx))

	require.NoError(t, applyMigrations(ctx, pool))
	require.NoError(t, truncateTables(ctx, pool))

	repo := NewUserRepo(pool)

	u := &domain.User{Name: "Test", Email: "test@example.com", PasswordHash: "hash", Role: domain.RoleUser}
	require.NoError(t, repo.Create(ctx, u))
	require.NotEmpty(t, u.ID)

	got, err := repo.GetByEmail(ctx, "test@example.com")
	require.NoError(t, err)
	require.Equal(t, u.ID, got.ID)
	require.Equal(t, "test@example.com", got.Email)

	require.NoError(t, truncateTables(ctx, pool))
}

func TestUserRepo_Integration_SQLExecution_NotFound(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)
	defer pool.Close()

	require.NoError(t, pool.Ping(ctx))

	require.NoError(t, applyMigrations(ctx, pool))
	require.NoError(t, truncateTables(ctx, pool))

	repo := NewUserRepo(pool)

	_, err = repo.GetByID(ctx, "missing")
	require.Error(t, err)
	require.ErrorIs(t, err, apperrors.ErrNotFound)

	require.NoError(t, truncateTables(ctx, pool))
}

func truncateTables(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, "TRUNCATE TABLE weather_history, cities, users")
	return err
}

func applyMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	b, err := os.ReadFile("../../../migrations/001_init.sql")
	if err != nil {
		return err
	}
	_, err = pool.Exec(ctx, string(b))
	return err
}
