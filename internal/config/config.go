package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Env   string
	HTTP  HTTPConfig
	DB    DBConfig
	API   APIConfig
	JWT   JWTConfig
	Redis RedisConfig
}

type HTTPConfig struct {
	Port            string
	ReadTimeoutSec  int
	WriteTimeoutSec int
	IdleTimeoutSec  int
}

type DBConfig struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime int
}

type APIConfig struct {
	WeatherAPIKey       string
	CountryStateCityKey string
}

type JWTConfig struct {
	Secret           string
	AccessTTLMinutes int
}

type RedisConfig struct {
	Addr string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Env: getEnv("APP_ENV", "production"),
		HTTP: HTTPConfig{
			Port:            getEnv("HTTP_PORT", "8080"),
			ReadTimeoutSec:  getEnvInt("HTTP_READ_TIMEOUT", 10),
			WriteTimeoutSec: getEnvInt("HTTP_WRITE_TIMEOUT", 30),
			IdleTimeoutSec:  getEnvInt("HTTP_IDLE_TIMEOUT", 60),
		},
		DB: DBConfig{
			DSN:             mustGetEnv("DATABASE_URL"),
			MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 10),
			ConnMaxLifetime: getEnvInt("DB_CONN_MAX_LIFETIME", 300),
		},
		API: APIConfig{
			WeatherAPIKey:       mustGetEnv("WEATHER_API_KEY"),
			CountryStateCityKey: mustGetEnv("COUNTRY_STATE_CITY_KEY"),
		},
		JWT: JWTConfig{
			Secret:           mustGetEnv("JWT_SECRET"),
			AccessTTLMinutes: getEnvInt("JWT_ACCESS_TTL_MINUTES", 60),
		},
		Redis: RedisConfig{
			Addr: getEnv("REDIS_ADDR", ""),
		},
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func mustGetEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("required environment variable %q is not set", key))
	}
	return v
}
