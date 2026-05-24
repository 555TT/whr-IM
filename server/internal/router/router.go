package router

import (
	"log"

	"whr-im/server/internal/config"
	"whr-im/server/internal/handler"
	"whr-im/server/internal/middleware"
	"whr-im/server/internal/repository"
	"whr-im/server/internal/service"
	"whr-im/server/internal/ws"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func New(cfg *config.Config) *gin.Engine {
	dsn := cfg.MySQL.DSN
	if dsn == "" {
		log.Fatal("mysql dsn is required")
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	userRepo, err := repository.NewGormUserRepository(db)
	if err != nil {
		log.Fatal(err)
	}
	friendRepo, err := repository.NewGormFriendRepository(db)
	if err != nil {
		log.Fatal(err)
	}
	messageRepo, err := repository.NewGormMessageRepository(db)
	if err != nil {
		log.Fatal(err)
	}

	return NewWithRepositories(userRepo, friendRepo, messageRepo)
}

func NewWithUserRepository(userRepo repository.UserRepository) *gin.Engine {
	return NewWithRepositories(userRepo, nil, nil)
}

func NewWithRepositories(userRepo repository.UserRepository, friendRepo repository.FriendRepository, messageRepo repository.MessageRepository) *gin.Engine {
	r := gin.Default()
	r.GET("/health", handler.Health)

	hub := ws.NewHub()
	authService := service.NewAuthService(userRepo, "dev-secret")
	authHandler := handler.NewAuthUserHandler(authService)
	wsHandler := handler.NewWebSocketHandler(authService, hub)

	r.GET("/ws", wsHandler.Connect)

	api := r.Group("/api")
	api.POST("/auth/register", authHandler.Register)
	api.POST("/auth/login", authHandler.Login)

	authed := api.Group("")
	authed.Use(middleware.Auth(authService))
	authed.GET("/users/me", authHandler.Me)
	authed.PUT("/users/me", authHandler.UpdateMe)

	if friendRepo != nil {
		friendService := service.NewFriendService(friendRepo, userRepo)
		friendHandler := handler.NewFriendHandler(friendService)
		authed.POST("/friend-requests", friendHandler.CreateRequest)
		authed.GET("/friend-requests/incoming", friendHandler.ListIncomingRequests)
		authed.PUT("/friend-requests/:id/accept", friendHandler.AcceptRequest)
		authed.PUT("/friend-requests/:id/reject", friendHandler.RejectRequest)
		authed.GET("/friends", friendHandler.ListFriends)

		if messageRepo != nil {
			messageService := service.NewMessageService(messageRepo, friendRepo, hub)
			messageHandler := handler.NewMessageHandler(messageService)
			authed.POST("/messages", messageHandler.Create)
			authed.GET("/messages", messageHandler.List)
		}
	}

	return r
}
