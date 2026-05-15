package main

import "os"

type Config struct {
	PostgresDSN string
	RedisAddr   string
	Port        string
	AiStudioKey string
}

func LoadConfig() Config {
	cfg := Config{
		PostgresDSN: lookupEnv("DATABASE_DSN", "postgres://dada:password@localhost:5432/dada?sslmode=disable"),
		RedisAddr:   lookupEnv("REDIS_ADDR", "localhost:6379"),
		Port:        lookupEnv("PORT", "8080"),
		AiStudioKey: os.Getenv("AISTUDIO_KEY"),
	}
	return cfg
}

func lookupEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
