package config

import (
	"log"
	"os"
	"strconv"

	"github.com/lpernett/godotenv"
)

type Config struct {
	DB_DSN        string
	JWTSecret     string
	Port          string
	ClientID      string
	ClientSecret  string
	CallbackURL   string
	ClientURL     string
	RedisADDR     string
	RedisPassword string
	SMTPHost      string
	SMTPPort      int
	SMTPUser      string
	SMTPPass      string
	SMTPFrom      string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Fatal(".env file is not exist or load incorrectly")
	}
	SMTPPort, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil {
		log.Fatal("Error with SMTP port")
	}
	cfg := &Config{
		DB_DSN:        os.Getenv("DB_DSN"),
		JWTSecret:     os.Getenv("JWT_SECRET"),
		Port:          os.Getenv("PORT"),
		ClientID:      os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret:  os.Getenv("GOOGLE_CLIENT_SECRET"),
		CallbackURL:   os.Getenv("CLIENT_CALLBACK_URL"),
		ClientURL:     os.Getenv("CLIENT_URL"),
		RedisADDR:     os.Getenv("REDIS_ADDR"),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
		SMTPHost:      os.Getenv("SMTP_HOST"),
		SMTPPort:      SMTPPort,
		SMTPUser:      os.Getenv("SMTP_USER"),
		SMTPPass:      os.Getenv("SMTP_PASSWORD"),
		SMTPFrom:      os.Getenv("SMTP_FROM"),
	}
	if cfg.Port == "" {
		cfg.Port = "8000"
	}

	if cfg.ClientURL == "" {
		cfg.ClientURL = "http://localhost:5173"
	}

	if cfg.CallbackURL == "" {
		cfg.CallbackURL = "http://localhost:" + cfg.Port + "/auth/google/callback"
	}

	if cfg.DB_DSN == "" || cfg.JWTSecret == "" || cfg.ClientID == "" || cfg.ClientSecret == "" {
		log.Fatal("All environment variables must be set")
	}

	return cfg
}
