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
}

func Load() *Config {
	_ = loadEnv()

	listenAddr := getEnv("LISTEN_ADDR", ":8080")
	steamAPIKey := os.Getenv("STEAM_API_KEY")
	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")

	if steamAPIKey == "" {
		log.Fatal("STEAM_API_KEY is not set")
	}

	return &Config{
		ListenAddr:  listenAddr,
		SteamAPIKey: steamAPIKey,
		RedisAddr:   redisAddr,
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
