package config

import "os"

type Config struct {
	// adresse d'ecoute du serveur
	HTTPAddr string
	// connection string de postgre
	DatabaseURL string
	// type d'environnement
	Env string
	// url API
	APIURL string
}

func Load() Config {
	return Config{
		HTTPAddr:    env("PGO_HTTP_ADDR", ":8080"),
		DatabaseURL: env("PGO_DATABASE_URL", "postgres://pgo:pgo@localhost:5433/pgo?sslmode=disable"),
		Env:         env("PGO_ENV", "dev"),
		APIURL:      env("PGO_API_URL", "http://localhost:8080"),
	}
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
