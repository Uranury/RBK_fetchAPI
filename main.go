package main

import (
	"context"
	"log"

	"github.com/Uranury/RBK_fetchAPI/config"
	_ "github.com/Uranury/RBK_fetchAPI/docs"
	"github.com/Uranury/RBK_fetchAPI/internal/server"
	"github.com/redis/go-redis/v9"
)

// TODO: Add more services/features, and swagger to every handler
// TODO: Write unit tests for services
// TODO: Add rate limiting
// TODO: Make docker and docker compose files
// TODO: Push to remote
// TODO: Add a README.md, .env.example, dockerignore
// TODO: Prepare a presentation

// @title           Steam API Wrapper
// @version         1.0
// @description     This is a service that resolves Steam vanity URLs and returns SteamIDs.
// @host            localhost:8080
// @BasePath        /
func main() {
	cfg := config.Load()

	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr,
		DB:   0,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		panic(err)
	}

	server, err := server.NewServer(cfg, rdb)
	if err != nil {
		log.Fatalf("Couldn't сreate server: %v", err)
	}
	log.Fatal(server.Start())
}
