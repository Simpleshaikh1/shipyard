// Package config reads settings from environment variables. Containers and
// Kubernetes both configure apps this way (12-factor), so we start here.
package config

import "os"

type Config struct {
	Addr     string // where the HTTP server listens, ":8080"
	LogLevel string // debug | info | warn | error
}

func Load() Config {
	return Config{
		Addr:     getenv("API_ADDR", ":8080"),
		LogLevel: getenv("LOG_LEVEL", "info"),
	}
}

func getenv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
