package config

import (
	"log"
	"os"
	"strconv"

	"github.com/lpernett/godotenv"
)

type Config struct {
	// DB
	DB_DSN string

	// AUTH
	JWTSecret    string
	Port         string
	ClientID     string
	ClientSecret string
	CallbackURL  string
	ClientURL    string
	ServerURL    string

	// REDIS
	RedisADDR     string
	RedisPassword string
	RedisTLS      bool

	// SMTP (legacy, kept for local dev)
	SMTPHost string
	SMTPPort int
	SMTPUser string
	SMTPPass string
	SMTPFrom string

	// Brevo API
	BrevoAPIKey string

	// AWS
	AWSRegion          string
	AWSS3Bucket        string
	AWSAccessKeyID     string
	AWSSecretAccessKey string
}

func Load() *Config {
	// Load .env file explicitly from the mounted volume path
	if os.Getenv("MODE") == "dev" {
		if err := godotenv.Load("/app/.env"); err != nil {
			log.Println("Error loading .env file:", err.Error())
			log.Fatal(".env file is not exist or load incorrectly")
		}
	}

	// Determine Redis keys based on environment mode
	redisAddrKey := "REDIS_ADDR"

	if os.Getenv("MODE") == "dev" {
		redisAddrKey = "REDIS_ADDR_DEV"
	}

	redisTLS := true
	redisTLSStr := os.Getenv("REDIS_TLS")
	if redisTLSStr != "true" {
		redisTLS = false
	}

	// Parse optional SMTP port (not required when using Brevo API)
	smtpPort, _ := strconv.Atoi(os.Getenv("SMTP_PORT"))

	// Initialize configuration struct using determined keys
	cfg := &Config{
		DB_DSN:       getEnvRequired("DB_DSN"),
		JWTSecret:    getEnvRequired("JWT_SECRET"),
		Port:         getEnv("PORT", "8000"),
		ClientID:     getEnvRequired("GOOGLE_CLIENT_ID"),
		ClientSecret: getEnvRequired("GOOGLE_CLIENT_SECRET"),
		CallbackURL:  getEnvRequired("CLIENT_CALLBACK_URL"),
		ClientURL:    getEnvRequired("CLIENT_URL"),
		ServerURL:    getEnvRequired("SERVER_URL"),

		RedisADDR:     getEnvRequired(redisAddrKey),
		RedisPassword: getEnvRequired("REDIS_PASSWORD"),
		RedisTLS:      redisTLS,

		SMTPHost: getEnv("SMTP_HOST", ""),
		SMTPPort: smtpPort,
		SMTPUser: getEnv("SMTP_USER", ""),
		SMTPPass: getEnv("SMTP_PASSWORD", ""),
		SMTPFrom: getEnv("SMTP_FROM", ""),

		BrevoAPIKey: getEnvRequired("BREVO_API_KEY"),

		AWSAccessKeyID:     getEnvRequired("AWS_ACCESS_KEY"),
		AWSSecretAccessKey: getEnvRequired("AWS_SECRET_ACCESS_KEY"),
		AWSS3Bucket:        getEnvRequired("AWS_S3_BUCKET"),
		AWSRegion:          getEnvRequired("AWS_REGION"),
	}

	return cfg
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvRequired(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("Required environment variable %s is not set", key)
	}
	return value
}
