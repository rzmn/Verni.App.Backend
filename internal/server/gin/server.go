package ginServer

import (
	"net/http"
	"time"
	"verni/internal/requestHandlers/accessToken"
	"verni/internal/requestHandlers/auth"
	"verni/internal/requestHandlers/avatars"
	"verni/internal/requestHandlers/friends"
	"verni/internal/requestHandlers/profile"
	"verni/internal/requestHandlers/spendings"
	"verni/internal/requestHandlers/users"
	"verni/internal/requestHandlers/verification"
	"verni/internal/schema"
	"verni/internal/server"
	"verni/internal/services/logging"
	"verni/internal/services/realtimeEvents"
	ginLongpollRealtimeEvents "verni/internal/services/realtimeEvents/longpoll"

	"github.com/gin-gonic/gin"
)

type RequestHandlers struct {
	Auth         auth.RequestsHandler
	Spendings    spendings.RequestsHandler
	Friends      friends.RequestsHandler
	Profile      profile.RequestsHandler
	Verification verification.RequestsHandler
	Users        users.RequestsHandler
	Avatars      avatars.RequestsHandler
}

type GinConfig struct {
	TimeoutSec     int    `json:"timeoutSec"`
	IdleTimeoutSec int    `json:"idleTimeoutSec"`
	RunMode        string `json:"runMode"`
	Port           string `json:"port"`
}

type ginAccessTokenChecker struct {
	handler     gin.HandlerFunc
	accessToken func(c *gin.Context) schema.UserId
}

const (
	accessTokenSubjectKey = "verni-subject"
)

func New(
	config GinConfig,
	accessTokenChecker accessToken.RequestHandler,
	requestHandlersBuilder func(realtimeEvents realtimeEvents.Service) RequestHandlers,
	logger logging.Service,
) server.Server {
	server := createGinServer(
		config,
		accessTokenChecker,
		requestHandlersBuilder,
		logger,
	)
	return &server
}

type ginServer struct {
	server http.Server
	logger logging.Service
}

func (c *ginServer) ListenAndServe() {
	c.logger.LogInfo("[info] start http server listening %s", c.server.Addr)
	c.server.ListenAndServe()
}

func createGinServer(
	config GinConfig,
	accessTokenChecker accessToken.RequestHandler,
	requestHandlersBuilder func(realtimeEvents realtimeEvents.Service) RequestHandlers,
	logger logging.Service,
) ginServer {
	logger.LogInfo("creating gin server with config %v", config)
	gin.SetMode(config.RunMode)
	router := gin.New()
	tokenChecker := ginAccessTokenChecker{
		handler: func(c *gin.Context) {
			accessTokenChecker.CheckToken(
				c.Request.Header.Get("Authorization"),
				func(code schema.StatusCode, response schema.Response[schema.UserId]) {
					c.Request.Header.Set(accessTokenSubjectKey, string(response.Response))
					c.Next()
				},
				func(code schema.StatusCode, error schema.Response[schema.Error]) {
					c.AbortWithStatusJSON(int(code), error)
				},
			)
		},
		accessToken: func(c *gin.Context) schema.UserId {
			return schema.UserId(c.Request.Header.Get(accessTokenSubjectKey))
		},
	}
	longpollService := ginLongpollRealtimeEvents.New(router, logger, tokenChecker.handler)
	handlers := requestHandlersBuilder(longpollService)
	{
		auth := router.Group("/auth")
		{
			auth.PUT("/signup", ginRequestHandler(func(c *gin.Context, request schema.SignupRequest) {
				handlers.Auth.Signup(request, ginSuccessResponse[schema.Response[schema.Session]](c), ginFailureResponse(c))
			}))
			auth.PUT("/login", ginRequestHandler(func(c *gin.Context, request schema.LoginRequest) {
				handlers.Auth.Login(request, ginSuccessResponse[schema.Response[schema.Session]](c), ginFailureResponse(c))
			}))
			auth.PUT("/refresh", tokenChecker.handler, ginRequestHandler(func(c *gin.Context, request schema.RefreshRequest) {
				handlers.Auth.Refresh(request, ginSuccessResponse[schema.Response[schema.Session]](c), ginFailureResponse(c))
			}))
			auth.PUT("/updateEmail", tokenChecker.handler, ginRequestHandler(func(c *gin.Context, request schema.UpdateEmailRequest) {
				subject := schema.UserId(tokenChecker.accessToken(c))
				handlers.Auth.UpdateEmail(subject, request, ginSuccessResponse[schema.Response[schema.Session]](c), ginFailureResponse(c))
			}))
			auth.PUT("/updatePassword", tokenChecker.handler, ginRequestHandler(func(c *gin.Context, request schema.UpdatePasswordRequest) {
				subject := schema.UserId(tokenChecker.accessToken(c))
				handlers.Auth.UpdatePassword(subject, request, ginSuccessResponse[schema.Response[schema.Session]](c), ginFailureResponse(c))
			}))
			auth.DELETE("/logout", tokenChecker.handler, func(c *gin.Context) {
				subject := schema.UserId(tokenChecker.accessToken(c))
				handlers.Auth.Logout(subject, ginSuccessResponse[schema.VoidResponse](c), ginFailureResponse(c))
			})
			auth.PUT("/registerForPushNotifications", tokenChecker.handler, ginRequestHandler(func(c *gin.Context, request schema.RegisterForPushNotificationsRequest) {
				subject := schema.UserId(tokenChecker.accessToken(c))
				handlers.Auth.RegisterForPushNotifications(subject, request, ginSuccessResponse[schema.VoidResponse](c), ginFailureResponse(c))
			}))
		}
		spendings := router.Group("/spendings", tokenChecker.handler)
		{
			spendings.POST("/addExpense", ginRequestHandler(func(c *gin.Context, request schema.AddExpenseRequest) {
				subject := schema.UserId(tokenChecker.accessToken(c))
				handlers.Spendings.AddExpense(subject, request, ginSuccessResponse[schema.Response[schema.IdentifiableExpense]](c), ginFailureResponse(c))
			}))
			spendings.POST("/removeExpense", ginRequestHandler(func(c *gin.Context, request schema.RemoveExpenseRequest) {
				subject := schema.UserId(tokenChecker.accessToken(c))
				handlers.Spendings.RemoveExpense(subject, request, ginSuccessResponse[schema.Response[schema.IdentifiableExpense]](c), ginFailureResponse(c))
			}))
			spendings.GET("/getBalance", func(c *gin.Context) {
				subject := schema.UserId(tokenChecker.accessToken(c))
				handlers.Spendings.GetBalance(subject, ginSuccessResponse[schema.Response[[]schema.Balance]](c), ginFailureResponse(c))
			})
			spendings.GET("/getExpenses", ginRequestHandler(func(c *gin.Context, request schema.GetExpensesRequest) {
				subject := schema.UserId(tokenChecker.accessToken(c))
				handlers.Spendings.GetExpenses(subject, request, ginSuccessResponse[schema.Response[[]schema.IdentifiableExpense]](c), ginFailureResponse(c))
			}))
			spendings.GET("/getExpense", ginRequestHandler(func(c *gin.Context, request schema.GetExpenseRequest) {
				subject := schema.UserId(tokenChecker.accessToken(c))
				handlers.Spendings.GetExpense(subject, request, ginSuccessResponse[schema.Response[schema.IdentifiableExpense]](c), ginFailureResponse(c))
			}))
		}
		friends := router.Group("/friends", tokenChecker.handler)
		{
			friends.POST("/acceptRequest", ginRequestHandler(func(c *gin.Context, request schema.AcceptFriendRequest) {
				subject := schema.UserId(tokenChecker.accessToken(c))
				handlers.Friends.AcceptRequest(subject, request, ginSuccessResponse[schema.VoidResponse](c), ginFailureResponse(c))
			}))
			friends.GET("/get", ginRequestHandler(func(c *gin.Context, request schema.GetFriendsRequest) {
				subject := schema.UserId(tokenChecker.accessToken(c))
				handlers.Friends.GetFriends(subject, request, ginSuccessResponse[schema.Response[map[schema.FriendStatus][]schema.UserId]](c), ginFailureResponse(c))
			}))
			friends.POST("/rejectRequest", ginRequestHandler(func(c *gin.Context, request schema.RejectFriendRequest) {
				subject := schema.UserId(tokenChecker.accessToken(c))
				handlers.Friends.RejectRequest(subject, request, ginSuccessResponse[schema.VoidResponse](c), ginFailureResponse(c))
			}))
			friends.POST("/rollbackRequest", ginRequestHandler(func(c *gin.Context, request schema.RollbackFriendRequest) {
				subject := schema.UserId(tokenChecker.accessToken(c))
				handlers.Friends.RollbackRequest(subject, request, ginSuccessResponse[schema.VoidResponse](c), ginFailureResponse(c))
			}))
			friends.POST("/sendRequest", ginRequestHandler(func(c *gin.Context, request schema.SendFriendRequest) {
				subject := schema.UserId(tokenChecker.accessToken(c))
				handlers.Friends.SendRequest(subject, request, ginSuccessResponse[schema.VoidResponse](c), ginFailureResponse(c))
			}))
			friends.POST("/unfriend", ginRequestHandler(func(c *gin.Context, request schema.UnfriendRequest) {
				subject := schema.UserId(tokenChecker.accessToken(c))
				handlers.Friends.Unfriend(subject, request, ginSuccessResponse[schema.VoidResponse](c), ginFailureResponse(c))
			}))
		}
		profile := router.Group("/profile", tokenChecker.handler)
		{
			profile.GET("/getInfo", func(c *gin.Context) {
				subject := schema.UserId(tokenChecker.accessToken(c))
				handlers.Profile.GetInfo(subject, ginSuccessResponse[schema.Response[schema.Profile]](c), ginFailureResponse(c))
			})
			profile.PUT("/setAvatar", ginRequestHandler(func(c *gin.Context, request schema.SetAvatarRequest) {
				subject := schema.UserId(tokenChecker.accessToken(c))
				handlers.Profile.SetAvatar(subject, request, ginSuccessResponse[schema.Response[schema.ImageId]](c), ginFailureResponse(c))
			}))
			profile.PUT("/setDisplayName", ginRequestHandler(func(c *gin.Context, request schema.SetDisplayNameRequest) {
				subject := schema.UserId(tokenChecker.accessToken(c))
				handlers.Profile.SetDisplayName(subject, request, ginSuccessResponse[schema.VoidResponse](c), ginFailureResponse(c))
			}))
		}
		verification := router.Group("/verification", tokenChecker.handler)
		{
			verification.PUT("/confirmEmail", ginRequestHandler(func(c *gin.Context, request schema.ConfirmEmailRequest) {
				subject := schema.UserId(tokenChecker.accessToken(c))
				handlers.Verification.ConfirmEmail(subject, request, ginSuccessResponse[schema.VoidResponse](c), ginFailureResponse(c))
			}))
			verification.PUT("/sendEmailConfirmationCode", func(c *gin.Context) {
				subject := schema.UserId(tokenChecker.accessToken(c))
				handlers.Verification.SendEmailConfirmationCode(subject, ginSuccessResponse[schema.VoidResponse](c), ginFailureResponse(c))
			})
		}
		users := router.Group("/users", tokenChecker.handler)
		{
			users.GET("/get", ginRequestHandler(func(c *gin.Context, request schema.GetUsersRequest) {
				subject := schema.UserId(tokenChecker.accessToken(c))
				handlers.Users.GetUsers(subject, request, ginSuccessResponse[schema.Response[[]schema.User]](c), ginFailureResponse(c))
			}))
		}
		avatars := router.Group("/avatars")
		{
			avatars.GET("/get", ginRequestHandler(func(c *gin.Context, request schema.GetAvatarsRequest) {
				handlers.Avatars.GetAvatars(request, ginSuccessResponse[schema.Response[map[schema.ImageId]schema.Image]](c), ginFailureResponse(c))
			}))
		}
	}
	return ginServer{
		server: http.Server{
			Addr:         ":" + config.Port,
			Handler:      router,
			ReadTimeout:  time.Second * time.Duration(config.IdleTimeoutSec),
			WriteTimeout: time.Second * time.Duration(config.IdleTimeoutSec),
		},
		logger: logger,
	}
}

func ginRequestHandler[R any](success func(*gin.Context, R)) func(c *gin.Context) {
	return func(c *gin.Context) {
		var request R
		if err := c.BindJSON(&request); err != nil {
			failure := ginFailureResponse(c)
			failure(http.StatusBadRequest, schema.Failure(err, schema.CodeBadRequest))
		} else {
			success(c, request)
		}
	}
}

func ginSuccessResponse[R any](c *gin.Context) func(status schema.StatusCode, response R) {
	return func(status schema.StatusCode, response R) {
		c.JSON(int(status), response)
	}
}

func ginFailureResponse(c *gin.Context) func(status schema.StatusCode, response schema.Response[schema.Error]) {
	return func(status schema.StatusCode, response schema.Response[schema.Error]) {
		c.AbortWithStatusJSON(int(status), response)
	}
}
