package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	RedisAddr      string
	ServerPort     string
	MailtrapAPIKey string
	MailtrapURL    string
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	return &Config{
		RedisAddr:      getEnv("REDIS_ADDR", "localhost:6379"),
		ServerPort:     getEnv("SERVER_PORT", ":8081"),
		MailtrapAPIKey: getEnv("MAILTRAP_API_KEY", ""),
		MailtrapURL:    getEnv("MAILTRAP_URL", ""),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
