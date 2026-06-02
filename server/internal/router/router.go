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
	groupRepo, err := repository.NewGormGroupRepository(db)
	if err != nil {
		log.Fatal(err)
	}
	groupMessageRepo, err := repository.NewGormGroupMessageRepository(db)
	if err != nil {
		log.Fatal(err)
	}
	momentRepo, err := repository.NewGormMomentRepository(db)
	if err != nil {
		log.Fatal(err)
	}
	storage, err := service.NewObjectStorageFromConfig(cfg.ObjectStorage)
	if err != nil {
		log.Fatal(err)
	}

	return newEngine(userRepo, friendRepo, messageRepo, groupRepo, groupMessageRepo, momentRepo, storage)
}

func NewWithUserRepository(userRepo repository.UserRepository) *gin.Engine {
	storage, err := service.NewObjectStorageFromConfig(config.ObjectStorageConfig{PublicBaseURL: "http://localhost:9000"})
	if err != nil {
		log.Fatal(err)
	}
	return newEngine(userRepo, nil, nil, nil, nil, nil, storage)
}

func NewWithRepositories(
	userRepo repository.UserRepository,
	friendRepo repository.FriendRepository,
	messageRepo repository.MessageRepository,
	groupRepo repository.GroupRepository,
	groupMessageRepo repository.GroupMessageRepository,
	momentRepo repository.MomentRepository,
) *gin.Engine {
	storage, err := service.NewObjectStorageFromConfig(config.ObjectStorageConfig{PublicBaseURL: "http://localhost:9000"})
	if err != nil {
		log.Fatal(err)
	}
	return newEngine(userRepo, friendRepo, messageRepo, groupRepo, groupMessageRepo, momentRepo, storage)
}

func newEngine(
	userRepo repository.UserRepository,
	friendRepo repository.FriendRepository,
	messageRepo repository.MessageRepository,
	groupRepo repository.GroupRepository,
	groupMessageRepo repository.GroupMessageRepository,
	momentRepo repository.MomentRepository,
	storage service.ObjectStorage,
) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.CORS())
	r.GET("/health", handler.Health)

	hub := ws.NewHub()
	authService := service.NewAuthService(userRepo, "dev-secret")
	authHandler := handler.NewAuthUserHandler(authService)
	wsHandler := handler.NewWebSocketHandler(authService, hub)
	uploadService := service.NewUploadService(storage)
	uploadHandler := handler.NewUploadHandler(uploadService)

	r.GET("/ws", wsHandler.Connect)

	api := r.Group("/api")
	api.POST("/auth/register", authHandler.Register)
	api.POST("/auth/login", authHandler.Login)

	authed := api.Group("")
	authed.Use(middleware.Auth(authService))
	authed.GET("/users/me", authHandler.Me)
	authed.PUT("/users/me", authHandler.UpdateMe)
	authed.PUT("/users/me/public-key", authHandler.UpdateMyPublicKey)
	authed.POST("/uploads/images", uploadHandler.UploadImage)

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

		if groupRepo != nil && groupMessageRepo != nil {
			groupService := service.NewGroupService(groupRepo, friendRepo, userRepo)
			groupMessageService := service.NewGroupMessageService(groupMessageRepo, groupRepo, hub)
			groupHandler := handler.NewGroupHandler(groupService)
			groupMessageHandler := handler.NewGroupMessageHandler(groupMessageService)

			groups := authed.Group("/groups")
			groups.POST("", groupHandler.Create)
			groups.GET("", groupHandler.ListMine)
			groups.GET("/:id", groupHandler.Detail)
			groups.POST("/:id/members", groupHandler.AddMembers)
			groups.DELETE("/:id/members/me", groupHandler.LeaveGroup)
			groups.POST("/:id/messages", groupMessageHandler.Create)
			groups.GET("/:id/messages", groupMessageHandler.List)
		}

		if momentRepo != nil {
			momentService := service.NewMomentService(momentRepo, friendRepo, userRepo, storage)
			momentHandler := handler.NewMomentHandler(momentService)
			authed.POST("/moments", momentHandler.Create)
			authed.GET("/moments", momentHandler.List)
			authed.DELETE("/moments/:id", momentHandler.Delete)
			authed.POST("/moments/:id/likes", momentHandler.Like)
			authed.DELETE("/moments/:id/likes/me", momentHandler.Unlike)
			authed.POST("/moments/:id/comments", momentHandler.CreateComment)
		}
	}

	return r
}
