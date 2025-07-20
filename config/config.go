package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ListenAddr  string
	SteamAPIKey string
	RedisAddr   string
	DB_URL      string
}

func Load() *Config {
	_ = loadEnv()

	listenAddr := getEnv("LISTEN_ADDR", ":8080")
	steamAPIKey := os.Getenv("STEAM_API_KEY")
	redisAddr := os.Getenv("REDIS_ADDR")
	db_url := os.Getenv("POSTGRES_DSN")

	if steamAPIKey == "" {
		log.Fatal("STEAM_API_KEY is not set")
	}

	return &Config{
		ListenAddr:  listenAddr,
		SteamAPIKey: steamAPIKey,
		RedisAddr:   redisAddr,
		DB_URL:      db_url,
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func loadEnv() error {
	return godotenv.Load()
}
