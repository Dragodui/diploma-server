package config

import (
	"log"
	"os"

	"github.com/lpernett/godotenv"
)

type Config struct {
	DB_DSN    string
	JWTSecret string
	Port      string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Fatal(".env file is not exist or load incorrectly")
	}
	cfg := &Config{
		DB_DSN:    os.Getenv("DB_DSN"),
		JWTSecret: os.Getenv("JWT_SECRET"),
		Port:      os.Getenv("PORT"),
	}
	if cfg.Port == "" {
		cfg.Port = "8000"
	}

	if cfg.DB_DSN == "" || cfg.JWTSecret == "" {
		log.Fatal("All environment variables must be set")
	}

	return cfg
}
