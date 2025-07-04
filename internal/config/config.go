package config

import (
	"log"
	"os"

	"github.com/lpernett/godotenv"
)

type Config struct {
	DB_DSN       string
	JWTSecret    string
	Port         string
	ClientID     string
	ClientSecret string
	CallbackURL  string
	ClientURL    string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Fatal(".env file is not exist or load incorrectly")
	}
	cfg := &Config{
		DB_DSN:       os.Getenv("DB_DSN"),
		JWTSecret:    os.Getenv("JWT_SECRET"),
		Port:         os.Getenv("PORT"),
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		CallbackURL:  os.Getenv("CLIENT_CALLBACK_URL"),
		ClientURL:    os.Getenv("CLIENT_URL"),
	}
	if cfg.Port == "" {
		cfg.Port = "8000"
	}

	if cfg.ClientURL == "" {
		cfg.ClientURL = "http://localhost:5173"
	}

	if cfg.CallbackURL == "" {
		cfg.CallbackURL = "http://locahost:" + cfg.Port + "/auth/google/callback"
	}

	if cfg.DB_DSN == "" || cfg.JWTSecret == "" || cfg.ClientID == "" || cfg.ClientSecret == "" {
		log.Fatal("All environment variables must be set")
	}

	return cfg
}
