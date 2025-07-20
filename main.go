package main

import (
	"context"
	"log"

	"github.com/Uranury/RBK_fetchAPI/config"
	_ "github.com/Uranury/RBK_fetchAPI/docs"
	"github.com/Uranury/RBK_fetchAPI/internal/server"
	"github.com/redis/go-redis/v9"
)

// @title           Steam API Wrapper
// @version         1.0
// @description     A lightweight service that integrates with the Steam Web API to fetch user profiles, owned games, and achievement data.
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
