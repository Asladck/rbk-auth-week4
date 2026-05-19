package handler

import (
	"bytes"
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
)

func TestAuthHandler_Register(t *testing.T) {
	log := zaptest.NewLogger(t)

	tcs := []struct {
		name           string
		body           string
		setupMock      func(s *mockdomain.MockAuthService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "invalid json",
			body:           "{",
			setupMock:      func(s *mockdomain.MockAuthService) {},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedError:  "body:",
		},
		{
			name:           "validation failure",
			body:           `{"name":"","email":"a@b.com","password":"secret123"}`,
			setupMock:      func(s *mockdomain.MockAuthService) {},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedError:  "body:",
		},
		{
			name: "happy path",
			body: `{"name":"Amir","email":"amir@example.com","password":"secret123"}`,
			setupMock: func(s *mockdomain.MockAuthService) {
				s.On("Register", mock.Anything, domain.RegisterInput{Name: "Amir", Email: "amir@example.com", Password: "secret123"}).Return(&domain.RegisterResponse{Message: "user created successfully"}, nil).Once()
			},
			expectedStatus: http.StatusCreated,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			svc := new(mockdomain.MockAuthService)
			h := NewAuthHandler(svc, log)

			r := chi.NewRouter()
			r.Post("/auth/register", h.Register)

			tc.setupMock(svc)

			req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(tc.body))
			req.Header.Set("Content-Type", "application/json")
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

			svc.AssertExpectations(t)
		})
	}
}
