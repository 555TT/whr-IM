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
	// GET /health: 健康检查接口，用于确认服务是否正常运行。
	r.GET("/health", handler.Health)

	hub := ws.NewHub()
	authService := service.NewAuthService(userRepo, "dev-secret")
	authHandler := handler.NewAuthUserHandler(authService)
	wsHandler := handler.NewWebSocketHandler(authService, hub)
	uploadService := service.NewUploadService(storage)
	uploadHandler := handler.NewUploadHandler(uploadService)

	// GET /ws: WebSocket 连接入口，用于建立实时消息通道。
	r.GET("/ws", wsHandler.Connect)

	api := r.Group("/api")
	// POST /api/auth/register: 用户注册接口，创建新账号并保存公钥信息。
	api.POST("/auth/register", authHandler.Register)
	// POST /api/auth/login: 用户登录接口，校验凭证后返回登录态。
	api.POST("/auth/login", authHandler.Login)

	authed := api.Group("")
	authed.Use(middleware.Auth(authService))
	// GET /api/users/me: 获取当前登录用户的个人信息。
	authed.GET("/users/me", authHandler.Me)
	// PUT /api/users/me: 更新当前登录用户的昵称、性别、个性签名等资料。
	authed.PUT("/users/me", authHandler.UpdateMe)
	// PUT /api/users/me/public-key: 更新当前登录用户的公钥信息，用于端到端加密通信。
	authed.PUT("/users/me/public-key", authHandler.UpdateMyPublicKey)
	// POST /api/uploads/images: 上传图片到对象存储，返回 objectKey 和可访问地址。
	authed.POST("/uploads/images", uploadHandler.UploadImage)

	if friendRepo != nil {
		friendService := service.NewFriendService(friendRepo, userRepo)
		friendHandler := handler.NewFriendHandler(friendService)
		// POST /api/friend-requests: 发送好友申请给目标用户。
		authed.POST("/friend-requests", friendHandler.CreateRequest)
		// GET /api/friend-requests/incoming: 获取当前用户收到的好友申请列表。
		authed.GET("/friend-requests/incoming", friendHandler.ListIncomingRequests)
		// PUT /api/friend-requests/:id/accept: 接受指定好友申请，并建立双向好友关系。
		authed.PUT("/friend-requests/:id/accept", friendHandler.AcceptRequest)
		// PUT /api/friend-requests/:id/reject: 拒绝指定好友申请。
		authed.PUT("/friend-requests/:id/reject", friendHandler.RejectRequest)
		// GET /api/friends: 获取当前用户的好友列表。
		authed.GET("/friends", friendHandler.ListFriends)

		if messageRepo != nil {
			messageService := service.NewMessageService(messageRepo, friendRepo, hub)
			messageHandler := handler.NewMessageHandler(messageService)
			// POST /api/messages: 发送单聊消息，保存消息并实时推送给接收方。
			authed.POST("/messages", messageHandler.Create)
			// GET /api/messages: 拉取当前用户与指定好友之间的单聊历史消息。
			authed.GET("/messages", messageHandler.List)
		}

		if groupRepo != nil && groupMessageRepo != nil {
			groupService := service.NewGroupService(groupRepo, friendRepo, userRepo)
			groupMessageService := service.NewGroupMessageService(groupMessageRepo, groupRepo, hub)
			groupHandler := handler.NewGroupHandler(groupService)
			groupMessageHandler := handler.NewGroupMessageHandler(groupMessageService)

			groups := authed.Group("/groups")
			// POST /api/groups: 创建群聊，并把创建者与指定好友加入群组。
			groups.POST("", groupHandler.Create)
			// GET /api/groups: 获取当前用户加入的群聊列表。
			groups.GET("", groupHandler.ListMine)
			// GET /api/groups/:id: 获取指定群聊的详情和成员信息。
			groups.GET("/:id", groupHandler.Detail)
			// POST /api/groups/:id/members: 向群聊中添加新的群成员。
			groups.POST("/:id/members", groupHandler.AddMembers)
			// DELETE /api/groups/:id/members/me: 当前用户退出指定群聊。
			groups.DELETE("/:id/members/me", groupHandler.LeaveGroup)
			// POST /api/groups/:id/messages: 发送群聊消息，保存后实时广播给群成员。
			groups.POST("/:id/messages", groupMessageHandler.Create)
			// GET /api/groups/:id/messages: 获取指定群聊的历史消息列表。
			groups.GET("/:id/messages", groupMessageHandler.List)
		}

		if momentRepo != nil {
			momentService := service.NewMomentService(momentRepo, friendRepo, userRepo, storage)
			momentHandler := handler.NewMomentHandler(momentService)
			// POST /api/moments: 发布一条朋友圈动态，可附带图片 objectKey 列表。
			authed.POST("/moments", momentHandler.Create)
			// GET /api/moments: 获取当前用户可见的朋友圈动态列表。
			authed.GET("/moments", momentHandler.List)
			// DELETE /api/moments/:id: 删除当前用户自己发布的朋友圈动态。
			authed.DELETE("/moments/:id", momentHandler.Delete)
			// POST /api/moments/:id/likes: 给指定朋友圈动态点赞。
			authed.POST("/moments/:id/likes", momentHandler.Like)
			// DELETE /api/moments/:id/likes/me: 取消当前用户对指定朋友圈动态的点赞。
			authed.DELETE("/moments/:id/likes/me", momentHandler.Unlike)
			// POST /api/moments/:id/comments: 给指定朋友圈动态发表评论。
			authed.POST("/moments/:id/comments", momentHandler.CreateComment)
		}
	}

	return r
}
