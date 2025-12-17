package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBUser string
	DBPass string
	DBHost string
	DBPort string
	DBName string
}

func getEnv(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return defaultValue
}

func Load() Config {
	_ = godotenv.Load()

	cfg := Config{
		DBUser: getEnv("DB_USER", "nuser"),
		DBPass: getEnv("DB_PASS", "npass"),
		DBHost: getEnv("DB_HOST", "localhost"),
		DBPort: getEnv("DB_PORT", "3306"),
		DBName: getEnv("DB_NAME", "ndb"),
	}

	if cfg.DBUser == "" {
		log.Fatal("DB_USER not set")
	}

	return cfg
}
