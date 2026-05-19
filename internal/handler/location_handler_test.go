package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"weather-auth-api/internal/domain"
	mockdomain "weather-auth-api/mocks/domain"
	"weather-auth-api/pkg/apperrors"
)

func TestLocationHandler_GetStates(t *testing.T) {
	log := zaptest.NewLogger(t)

	tcs := []struct {
		name           string
		country        string
		setupMock      func(s *mockdomain.MockLocationService)
		expectedStatus int
		expectedData   any
		expectedError  string
	}{
		{
			name:    "not found",
			country: "KZ",
			setupMock: func(s *mockdomain.MockLocationService) {
				s.On("GetStates", mock.Anything, "KZ").Return(([]domain.State)(nil), apperrors.ErrNotFound).Once()
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "resource not found",
		},
		{
			name:    "internal server error",
			country: "KZ",
			setupMock: func(s *mockdomain.MockLocationService) {
				s.On("GetStates", mock.Anything, "KZ").Return(([]domain.State)(nil), assert.AnError).Once()
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "internal server error",
		},
		{
			name:    "happy path",
			country: "KZ",
			setupMock: func(s *mockdomain.MockLocationService) {
				s.On("GetStates", mock.Anything, "KZ").Return([]domain.State{{Name: "Almaty", StateCode: "ALA"}}, nil).Once()
			},
			expectedStatus: http.StatusOK,
			expectedData:   []any{map[string]any{"name": "Almaty", "iso2": "ALA"}},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			svc := new(mockdomain.MockLocationService)
			h := NewLocationHandler(svc, log)

			r := chi.NewRouter()
			r.Get("/locations/countries/{country}/states", h.GetStates)

			tc.setupMock(svc)

			req := httptest.NewRequest(http.MethodGet, "/locations/countries/"+tc.country+"/states", nil)
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)

			var out map[string]any
			require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &out))

			if tc.expectedError != "" {
				errVal, ok := out["error"].(string)
				require.True(t, ok)
				assert.Contains(t, errVal, tc.expectedError)
			}

			if tc.expectedData != nil {
				assert.Equal(t, tc.expectedData, out["data"])
			}

			svc.AssertExpectations(t)
		})
	}
}
