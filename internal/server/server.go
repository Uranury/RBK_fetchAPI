package server

import (
	"log"

	"github.com/Uranury/RBK_fetchAPI/config"
	"github.com/Uranury/RBK_fetchAPI/internal/handlers"
	"github.com/Uranury/RBK_fetchAPI/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Server struct {
	router      *gin.Engine
	cfg         *config.Config
	redisClient *redis.Client
	userHandler *handlers.UserHandler
}

func NewServer(cfg *config.Config, redisClient *redis.Client) (*Server, error) {
	steamService := services.NewSteamService(cfg.SteamAPIKey, redisClient)
	userHandler := handlers.NewUserHandler(steamService)

	server := &Server{
		router:      gin.Default(),
		cfg:         cfg,
		redisClient: redisClient,
		userHandler: userHandler,
	}

	server.setupRoutes()
	return server, nil
}

func (s *Server) Start() error {
	log.Printf("Listening on port %s...", s.cfg.ListenAddr)
	return s.router.Run(s.cfg.ListenAddr)
}

func (s *Server) setupRoutes() {
	s.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	s.router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"msg": "pong"})
	})
	s.router.GET("/steam_id", s.userHandler.GetVanityProfile)
}
