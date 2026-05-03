package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"weather-auth-api/internal/domain"
)

type CityHandler struct {
	svc domain.CityService
	log *zap.Logger
}

func NewCityHandler(svc domain.CityService, log *zap.Logger) *CityHandler {
	return &CityHandler{svc: svc, log: log}
}

func (h *CityHandler) Add(w http.ResponseWriter, r *http.Request) {
	authUser := domain.MustUserFromContext(r.Context())

	var in domain.AddCityInput
	if err := decodeJSON(r, &in); err != nil {
		writeError(w, h.log, err)
		return
	}

	city, err := h.svc.Add(r.Context(), authUser.ID, in)
	if err != nil {
		writeError(w, h.log, err)
		return
	}
	writeJSON(w, http.StatusCreated, city)
}

func (h *CityHandler) List(w http.ResponseWriter, r *http.Request) {
	authUser := domain.MustUserFromContext(r.Context())

	cities, err := h.svc.List(r.Context(), authUser.ID)
	if err != nil {
		writeError(w, h.log, err)
		return
	}
	writeJSON(w, http.StatusOK, cities)
}

func (h *CityHandler) Delete(w http.ResponseWriter, r *http.Request) {
	authUser := domain.MustUserFromContext(r.Context())
	cityID := chi.URLParam(r, "city_id")

	if err := h.svc.Delete(r.Context(), authUser.ID, cityID); err != nil {
		writeError(w, h.log, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
