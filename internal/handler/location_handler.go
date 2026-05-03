package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"weather-auth-api/internal/domain"
	"weather-auth-api/pkg/apperrors"
)

type LocationHandler struct {
	svc domain.LocationService
	log *zap.Logger
}

func NewLocationHandler(svc domain.LocationService, log *zap.Logger) *LocationHandler {
	return &LocationHandler{svc: svc, log: log}
}

func (h *LocationHandler) GetStates(w http.ResponseWriter, r *http.Request) {
	country := chi.URLParam(r, "country")
	if country == "" {
		writeError(w, h.log, &apperrors.ValidationError{
			Field:   "country",
			Message: "required path parameter",
		})
		return
	}

	states, err := h.svc.GetStates(r.Context(), country)
	if err != nil {
		writeError(w, h.log, err)
		return
	}
	writeJSON(w, http.StatusOK, states)
}
