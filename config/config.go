package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	JWTSecret   string
}

func LoadConfig() Config {
	if err := godotenv.Load(); err != nil {
		log.Fatal("error loading environment")
	}
	return Config{
		DatabaseURL: getEnv("DATABASE_URL"),
		JWTSecret:   getEnv("JWT_SECRET"),
	}
}

func getEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("%s is not set", key)
	}
	return value
}
