package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chiMW "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	"weather-auth-api/internal/domain"
	appMW "weather-auth-api/internal/middleware"
	"weather-auth-api/pkg/jwtutil"
)

func NewRouter(
	log *zap.Logger,
	jwtManager *jwtutil.Manager,
	authH *AuthHandler,
	userH *UserHandler,
	cityH *CityHandler,
	weatherH *WeatherHandler,
	locationH *LocationHandler,
) http.Handler {
	r := chi.NewRouter()

	r.Use(chiMW.RequestID)
	r.Use(chiMW.RealIP)
	r.Use(appMW.Recoverer(log))
	r.Use(appMW.Logger(log))
	r.Use(chiMW.Compress(5))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", authH.Register)
		r.Post("/login", authH.Login)
	})

	r.Group(func(r chi.Router) {
		r.Use(appMW.Auth(jwtManager, log))

		r.Get("/users/me", userH.Me)

		r.Route("/cities", func(r chi.Router) {
			r.Post("/", cityH.Add)
			r.Get("/", cityH.List)
			r.Delete("/{city_id}", cityH.Delete)
		})

		r.Get("/weather", weatherH.GetWeather)
		r.Get("/weather/history", weatherH.GetHistory)

		r.Get("/locations/countries/{country}/states", locationH.GetStates)

		r.Group(func(r chi.Router) {
			r.Use(appMW.RequireRole(domain.RoleAdmin))

			r.Get("/users", userH.List)
			r.Get("/users/{id}", userH.GetByID)
			r.Delete("/users/{id}", userH.Delete)
		})
	})

	return r
}
