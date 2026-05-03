package handler

import (
	"net/http"

	"go.uber.org/zap"

	"weather-auth-api/internal/domain"
)

type AuthHandler struct {
	svc domain.AuthService
	log *zap.Logger
}

func NewAuthHandler(svc domain.AuthService, log *zap.Logger) *AuthHandler {
	return &AuthHandler{svc: svc, log: log}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var in domain.RegisterInput
	if err := decodeJSON(r, &in); err != nil {
		writeError(w, h.log, err)
		return
	}

	resp, err := h.svc.Register(r.Context(), in)
	if err != nil {
		writeError(w, h.log, err)
		return
	}
	writeJSON(w, http.StatusCreated, resp)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var in domain.LoginInput
	if err := decodeJSON(r, &in); err != nil {
		writeError(w, h.log, err)
		return
	}

	token, err := h.svc.Login(r.Context(), in)
	if err != nil {
		writeError(w, h.log, err)
		return
	}
	writeJSON(w, http.StatusOK, token)
}
