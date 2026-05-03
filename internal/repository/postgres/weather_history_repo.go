package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"weather-auth-api/internal/domain"
)

type WeatherHistoryRepo struct {
	db *pgxpool.Pool
}

func NewWeatherHistoryRepo(db *pgxpool.Pool) *WeatherHistoryRepo {
	return &WeatherHistoryRepo{db: db}
}

func (r *WeatherHistoryRepo) Save(ctx context.Context, h *domain.WeatherHistory) error {
	h.ID = uuid.NewString()
	h.RequestedAt = time.Now().UTC()

	q := `
		INSERT INTO weather_history
			(id, user_id, city, temp_c, feels_like, humidity, wind_kph, condition, requested_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`

	_, err := r.db.Exec(ctx, q,
		h.ID, h.UserID, h.City,
		h.TempC, h.FeelsLike, h.Humidity, h.WindKph,
		h.Condition, h.RequestedAt,
	)
	if err != nil {
		return fmt.Errorf("WeatherHistoryRepo.Save: %w", err)
	}
	return nil
}

// List returns history for a user filtered by city (case-insensitive), newest first.
// limit=0 means unlimited.
func (r *WeatherHistoryRepo) List(ctx context.Context, userID string, filter domain.HistoryFilter) ([]*domain.WeatherHistory, error) {
	var sb strings.Builder
	args := []any{userID, filter.City}

	sb.WriteString(`
		SELECT id, user_id, city, temp_c, feels_like, humidity, wind_kph, condition, requested_at
		FROM weather_history
		WHERE user_id=$1 AND LOWER(city)=LOWER($2)
		ORDER BY requested_at DESC`)

	if filter.Limit > 0 {
		args = append(args, filter.Limit)
		fmt.Fprintf(&sb, "\nLIMIT $%d", len(args))
	}

	rows, err := r.db.Query(ctx, sb.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("WeatherHistoryRepo.List: %w", err)
	}
	defer rows.Close()

	var history []*domain.WeatherHistory
	for rows.Next() {
		h := &domain.WeatherHistory{}
		if err := rows.Scan(
			&h.ID, &h.UserID, &h.City,
			&h.TempC, &h.FeelsLike, &h.Humidity, &h.WindKph,
			&h.Condition, &h.RequestedAt,
		); err != nil {
			return nil, fmt.Errorf("WeatherHistoryRepo.List scan: %w", err)
		}
		history = append(history, h)
	}
	return history, rows.Err()
}
