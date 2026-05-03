package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"weather-auth-api/internal/client"
	"weather-auth-api/internal/config"
	"weather-auth-api/internal/handler"
	"weather-auth-api/internal/repository/postgres"
	"weather-auth-api/internal/service"
	"weather-auth-api/pkg/jwtutil"
	"weather-auth-api/pkg/logger"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config.Load: %w", err)
	}

	log, err := logger.New(cfg.Env)
	if err != nil {
		return fmt.Errorf("logger.New: %w", err)
	}
	defer log.Sync()

	log.Info("starting server",
		zap.String("env", cfg.Env),
		zap.String("port", cfg.HTTP.Port),
	)

	ctx := context.Background()
	pool, err := postgres.NewPool(ctx, cfg.DB)
	if err != nil {
		return fmt.Errorf("postgres.NewPool: %w", err)
	}
	defer pool.Close()
	log.Info("database connected")

	userRepo := postgres.NewUserRepo(pool)
	cityRepo := postgres.NewCityRepo(pool)
	historyRepo := postgres.NewWeatherHistoryRepo(pool)

	weatherClient := client.NewWeatherAPIClient(cfg.API.WeatherAPIKey)
	locationClient := client.NewLocationAPIClient(cfg.API.CountryStateCityKey)

	jwtManager := jwtutil.NewManager(cfg.JWT.Secret, cfg.JWT.AccessTTLMinutes)

	authSvc := service.NewAuthService(userRepo, jwtManager)
	userSvc := service.NewUserService(userRepo)
	citySvc := service.NewCityService(cityRepo, userRepo)
	weatherSvc := service.NewWeatherService(userRepo, cityRepo, historyRepo, weatherClient, log)
	locationSvc := service.NewLocationService(locationClient)

	authH := handler.NewAuthHandler(authSvc, log)
	userH := handler.NewUserHandler(userSvc, log)
	cityH := handler.NewCityHandler(citySvc, log)
	weatherH := handler.NewWeatherHandler(weatherSvc, log)
	locationH := handler.NewLocationHandler(locationSvc, log)

	router := handler.NewRouter(log, jwtManager, authH, userH, cityH, weatherH, locationH)

	srv := &http.Server{
		Addr:         ":" + cfg.HTTP.Port,
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.HTTP.ReadTimeoutSec) * time.Second,
		WriteTimeout: time.Duration(cfg.HTTP.WriteTimeoutSec) * time.Second,
		IdleTimeout:  time.Duration(cfg.HTTP.IdleTimeoutSec) * time.Second,
	}

	shutdownErr := make(chan error, 1)
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		sig := <-quit
		log.Info("shutdown signal received", zap.String("signal", sig.String()))

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		shutdownErr <- srv.Shutdown(shutdownCtx)
	}()

	log.Info("http server listening", zap.String("addr", srv.Addr))

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("ListenAndServe: %w", err)
	}

	if err := <-shutdownErr; err != nil {
		return fmt.Errorf("graceful shutdown: %w", err)
	}

	log.Info("server stopped gracefully")
	return nil
}
