package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"go.uber.org/zap"

	"weather-auth-api/pkg/apperrors"
)

type envelope struct {
	Data  any    `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(envelope{Data: data})
}

func writeError(w http.ResponseWriter, log *zap.Logger, err error) {
	status, msg := mapError(err)
	if status >= 500 {
		log.Error("internal error", zap.Error(err))
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(envelope{Error: msg})
}

func mapError(err error) (int, string) {
	switch {
	case errors.Is(err, apperrors.ErrInvalidToken):
		return http.StatusUnauthorized, "invalid or expired token"
	case errors.Is(err, apperrors.ErrUnauthorized):
		return http.StatusUnauthorized, "unauthorized"
	case errors.Is(err, apperrors.ErrForbidden):
		return http.StatusForbidden, "forbidden"
	case errors.Is(err, apperrors.ErrBadCreds):
		return http.StatusUnauthorized, "invalid credentials"
	case errors.Is(err, apperrors.ErrNotFound):
		return http.StatusNotFound, err.Error()
	case errors.Is(err, apperrors.ErrAlreadyExists):
		return http.StatusConflict, err.Error()
	case errors.Is(err, apperrors.ErrInvalidInput):
		return http.StatusUnprocessableEntity, err.Error()
	case errors.Is(err, apperrors.ErrUserDeleted):
		return http.StatusGone, "user has been deleted"
	case errors.Is(err, apperrors.ErrExternalAPI):
		return http.StatusBadGateway, "upstream service error"
	default:
		return http.StatusInternalServerError, "internal server error"
	}
}

func decodeJSON(r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(nil, r.Body, 1<<20)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return &apperrors.ValidationError{Field: "body", Message: err.Error()}
	}
	return nil
}
