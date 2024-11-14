package server

import (
	"net/http"
	"time"
	"verni/internal/common"
	"verni/internal/requestHandlers/accessToken"
	"verni/internal/requestHandlers/auth"
	"verni/internal/requestHandlers/avatars"
	"verni/internal/requestHandlers/friends"
	"verni/internal/requestHandlers/profile"
	"verni/internal/requestHandlers/spendings"
	"verni/internal/requestHandlers/users"
	"verni/internal/requestHandlers/verification"
	"verni/internal/schema"
	"verni/internal/services/logging"
	"verni/internal/services/longpoll"

	"github.com/gin-gonic/gin"
)

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
	requestHandlersBuilder func(longpoll longpoll.Service) RequestHandlers,
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
	longpollService := longpoll.GinService(router, logger, tokenChecker.handler)
	handlers := requestHandlersBuilder(longpollService)
	registerRoutesAuth(router, tokenChecker, handlers.Auth)
	registerRoutesSpendings(router, tokenChecker, handlers.Spendings)
	registerRoutesFriends(router, tokenChecker, handlers.Friends)
	registerRoutesProfile(router, tokenChecker, handlers.Profile)
	registerRoutesVerification(router, tokenChecker, handlers.Verification)
	registerRoutesUsers(router, tokenChecker, handlers.Users)
	registerRoutesAvatars(router, handlers.Avatars)
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
			failure(http.StatusBadRequest, schema.Failure(common.NewErrorWithDescriptionValue(schema.CodeBadRequest, err.Error())))
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

func registerRoutesAuth(router *gin.Engine, tokenChecker ginAccessTokenChecker, handler auth.RequestsHandler) {
	router.PUT("/auth/signup", ginRequestHandler(func(c *gin.Context, request auth.SignupRequest) {
		handler.Signup(request, ginSuccessResponse[schema.Response[schema.Session]](c), ginFailureResponse(c))
	}))
	router.PUT("/auth/login", ginRequestHandler(func(c *gin.Context, request auth.LoginRequest) {
		handler.Login(request, ginSuccessResponse[schema.Response[schema.Session]](c), ginFailureResponse(c))
	}))
	router.PUT("/auth/refresh", tokenChecker.handler, ginRequestHandler(func(c *gin.Context, request auth.RefreshRequest) {
		handler.Refresh(request, ginSuccessResponse[schema.Response[schema.Session]](c), ginFailureResponse(c))
	}))
	router.PUT("/auth/updateEmail", tokenChecker.handler, ginRequestHandler(func(c *gin.Context, request auth.UpdateEmailRequest) {
		subject := schema.UserId(tokenChecker.accessToken(c))
		handler.UpdateEmail(subject, request, ginSuccessResponse[schema.Response[schema.Session]](c), ginFailureResponse(c))
	}))
	router.PUT("/auth/updatePassword", tokenChecker.handler, ginRequestHandler(func(c *gin.Context, request auth.UpdatePasswordRequest) {
		subject := schema.UserId(tokenChecker.accessToken(c))
		handler.UpdatePassword(subject, request, ginSuccessResponse[schema.Response[schema.Session]](c), ginFailureResponse(c))
	}))
	router.DELETE("/auth/logout", tokenChecker.handler, func(c *gin.Context) {
		subject := schema.UserId(tokenChecker.accessToken(c))
		handler.Logout(subject, ginSuccessResponse[schema.VoidResponse](c), ginFailureResponse(c))
	})
	router.PUT("/auth/registerForPushNotifications", tokenChecker.handler, ginRequestHandler(func(c *gin.Context, request auth.RegisterForPushNotificationsRequest) {
		subject := schema.UserId(tokenChecker.accessToken(c))
		handler.RegisterForPushNotifications(subject, request, ginSuccessResponse[schema.VoidResponse](c), ginFailureResponse(c))
	}))
}

func registerRoutesSpendings(router *gin.Engine, tokenChecker ginAccessTokenChecker, handler spendings.RequestsHandler) {
	group := router.Group("/spendings", tokenChecker.handler)
	group.POST("/addExpense", ginRequestHandler(func(c *gin.Context, request spendings.AddExpenseRequest) {
		subject := schema.UserId(tokenChecker.accessToken(c))
		handler.AddExpense(subject, request, ginSuccessResponse[schema.Response[schema.IdentifiableExpense]](c), ginFailureResponse(c))
	}))
	group.POST("/removeExpense", ginRequestHandler(func(c *gin.Context, request spendings.RemoveExpenseRequest) {
		subject := schema.UserId(tokenChecker.accessToken(c))
		handler.RemoveExpense(subject, request, ginSuccessResponse[schema.Response[schema.IdentifiableExpense]](c), ginFailureResponse(c))
	}))
	group.GET("/getBalance", func(c *gin.Context) {
		subject := schema.UserId(tokenChecker.accessToken(c))
		handler.GetBalance(subject, ginSuccessResponse[schema.Response[[]schema.Balance]](c), ginFailureResponse(c))
	})
	group.GET("/getExpenses", ginRequestHandler(func(c *gin.Context, request spendings.GetExpensesRequest) {
		subject := schema.UserId(tokenChecker.accessToken(c))
		handler.GetExpenses(subject, request, ginSuccessResponse[schema.Response[[]schema.IdentifiableExpense]](c), ginFailureResponse(c))
	}))
	group.GET("/getExpense", ginRequestHandler(func(c *gin.Context, request spendings.GetExpenseRequest) {
		subject := schema.UserId(tokenChecker.accessToken(c))
		handler.GetExpense(subject, request, ginSuccessResponse[schema.Response[schema.IdentifiableExpense]](c), ginFailureResponse(c))
	}))
}

func registerRoutesFriends(router *gin.Engine, tokenChecker ginAccessTokenChecker, handler friends.RequestsHandler) {
	group := router.Group("/friends", tokenChecker.handler)
	group.POST("/acceptRequest", ginRequestHandler(func(c *gin.Context, request friends.AcceptFriendRequest) {
		subject := schema.UserId(tokenChecker.accessToken(c))
		handler.AcceptRequest(subject, request, ginSuccessResponse[schema.VoidResponse](c), ginFailureResponse(c))
	}))
	group.GET("/get", ginRequestHandler(func(c *gin.Context, request friends.GetFriendsRequest) {
		subject := schema.UserId(tokenChecker.accessToken(c))
		handler.GetFriends(subject, request, ginSuccessResponse[schema.Response[map[schema.FriendStatus][]schema.UserId]](c), ginFailureResponse(c))
	}))
	group.POST("/rejectRequest", ginRequestHandler(func(c *gin.Context, request friends.RejectFriendRequest) {
		subject := schema.UserId(tokenChecker.accessToken(c))
		handler.RejectRequest(subject, request, ginSuccessResponse[schema.VoidResponse](c), ginFailureResponse(c))
	}))
	group.POST("/rollbackRequest", ginRequestHandler(func(c *gin.Context, request friends.RollbackFriendRequest) {
		subject := schema.UserId(tokenChecker.accessToken(c))
		handler.RollbackRequest(subject, request, ginSuccessResponse[schema.VoidResponse](c), ginFailureResponse(c))
	}))
	group.POST("/sendRequest", ginRequestHandler(func(c *gin.Context, request friends.SendFriendRequest) {
		subject := schema.UserId(tokenChecker.accessToken(c))
		handler.SendRequest(subject, request, ginSuccessResponse[schema.VoidResponse](c), ginFailureResponse(c))
	}))
	group.POST("/unfriend", ginRequestHandler(func(c *gin.Context, request friends.UnfriendRequest) {
		subject := schema.UserId(tokenChecker.accessToken(c))
		handler.Unfriend(subject, request, ginSuccessResponse[schema.VoidResponse](c), ginFailureResponse(c))
	}))
}

func registerRoutesProfile(router *gin.Engine, tokenChecker ginAccessTokenChecker, handler profile.RequestsHandler) {
	group := router.Group("/profile", tokenChecker.handler)
	group.GET("/getInfo", func(c *gin.Context) {
		subject := schema.UserId(tokenChecker.accessToken(c))
		handler.GetInfo(subject, ginSuccessResponse[schema.Response[schema.Profile]](c), ginFailureResponse(c))
	})
	group.PUT("/setAvatar", ginRequestHandler(func(c *gin.Context, request profile.SetAvatarRequest) {
		subject := schema.UserId(tokenChecker.accessToken(c))
		handler.SetAvatar(subject, request, ginSuccessResponse[schema.Response[schema.ImageId]](c), ginFailureResponse(c))
	}))
	group.PUT("/setDisplayName", ginRequestHandler(func(c *gin.Context, request profile.SetDisplayNameRequest) {
		subject := schema.UserId(tokenChecker.accessToken(c))
		handler.SetDisplayName(subject, request, ginSuccessResponse[schema.VoidResponse](c), ginFailureResponse(c))
	}))
}

func registerRoutesVerification(router *gin.Engine, tokenChecker ginAccessTokenChecker, handler verification.RequestsHandler) {
	group := router.Group("/verification", tokenChecker.handler)
	group.PUT("/confirmEmail", ginRequestHandler(func(c *gin.Context, request verification.ConfirmEmailRequest) {
		subject := schema.UserId(tokenChecker.accessToken(c))
		handler.ConfirmEmail(subject, request, ginSuccessResponse[schema.VoidResponse](c), ginFailureResponse(c))
	}))
	group.PUT("/sendEmailConfirmationCode", func(c *gin.Context) {
		subject := schema.UserId(tokenChecker.accessToken(c))
		handler.SendEmailConfirmationCode(subject, ginSuccessResponse[schema.VoidResponse](c), ginFailureResponse(c))
	})
}

func registerRoutesUsers(router *gin.Engine, tokenChecker ginAccessTokenChecker, handler users.RequestsHandler) {
	group := router.Group("/users", tokenChecker.handler)
	group.GET("/get", ginRequestHandler(func(c *gin.Context, request users.GetUsersRequest) {
		subject := schema.UserId(tokenChecker.accessToken(c))
		handler.GetUsers(subject, request, ginSuccessResponse[schema.Response[[]schema.User]](c), ginFailureResponse(c))
	}))
}

func registerRoutesAvatars(router *gin.Engine, handler avatars.RequestsHandler) {
	group := router.Group("/avatars")
	group.GET("/get", ginRequestHandler(func(c *gin.Context, request avatars.GetAvatarsRequest) {
		handler.GetAvatars(request, ginSuccessResponse[schema.Response[map[schema.ImageId]schema.Image]](c), ginFailureResponse(c))
	}))
}
