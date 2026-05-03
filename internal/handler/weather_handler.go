package handler

import (
	"net/http"
	"strconv"

	"go.uber.org/zap"

	"weather-auth-api/internal/domain"
	"weather-auth-api/pkg/apperrors"
)

// WeatherHandler wires HTTP to domain.WeatherService.
type WeatherHandler struct {
	svc domain.WeatherService
	log *zap.Logger
}

func NewWeatherHandler(svc domain.WeatherService, log *zap.Logger) *WeatherHandler {
	return &WeatherHandler{svc: svc, log: log}
}

// GetWeather handles GET /weather
func (h *WeatherHandler) GetWeather(w http.ResponseWriter, r *http.Request) {
	authUser := domain.MustUserFromContext(r.Context())

	resp, err := h.svc.GetWeatherForUser(r.Context(), authUser.ID)
	if err != nil {
		writeError(w, h.log, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// GetHistory handles GET /weather/history?city=Almaty&limit=10
func (h *WeatherHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	authUser := domain.MustUserFromContext(r.Context())

	city := r.URL.Query().Get("city")
	if city == "" {
		writeError(w, h.log, &apperrors.ValidationError{
			Field:   "city",
			Message: "required query parameter",
		})
		return
	}

	filter := domain.HistoryFilter{City: city}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 {
			writeError(w, h.log, &apperrors.ValidationError{
				Field:   "limit",
				Message: "must be a positive integer",
			})
			return
		}
		filter.Limit = limit
	}

	history, err := h.svc.GetHistory(r.Context(), authUser.ID, filter)
	if err != nil {
		writeError(w, h.log, err)
		return
	}
	writeJSON(w, http.StatusOK, history)
}
