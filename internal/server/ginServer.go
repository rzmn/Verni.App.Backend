package server

import (
	"net/http"
	"time"
	"verni/internal/common"
	httpserver "verni/internal/http-server"
	"verni/internal/requestHandlers/accessToken"
	"verni/internal/requestHandlers/auth"
	"verni/internal/requestHandlers/avatars"
	"verni/internal/requestHandlers/friends"
	"verni/internal/requestHandlers/profile"
	"verni/internal/requestHandlers/spendings"
	"verni/internal/requestHandlers/users"
	"verni/internal/requestHandlers/verification"
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
				func(code httpserver.StatusCode, response httpserver.Response[httpserver.UserId]) {
					c.Request.Header.Set(accessTokenSubjectKey, string(response.Response))
					c.Next()
				},
				func(code httpserver.StatusCode, error httpserver.Response[httpserver.Error]) {
					c.AbortWithStatusJSON(int(code), error)
				},
			)
		},
		accessToken: func(c *gin.Context) httpserver.UserId {
			return httpserver.UserId(c.Request.Header.Get(accessTokenSubjectKey))
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

func registerRoutesAuth(router *gin.Engine, tokenChecker ginAccessTokenChecker, handler auth.RequestsHandler) {
	router.PUT("/auth/signup", func(c *gin.Context) {
		var request auth.SignupRequest
		if err := c.BindJSON(&request); err != nil {
			c.AbortWithStatusJSON(
				http.StatusBadRequest,
				httpserver.Failure(common.NewErrorWithDescriptionValue(httpserver.CodeBadRequest, err.Error())),
			)
			return
		}
		handler.Signup(
			request,
			func(status httpserver.StatusCode, response httpserver.Response[httpserver.Session]) {
				c.JSON(int(status), response)
			},
			func(status httpserver.StatusCode, response httpserver.Response[httpserver.Error]) {
				c.AbortWithStatusJSON(int(status), response)
			},
		)
	})
	router.PUT("/auth/login", func(c *gin.Context) {
		var request auth.LoginRequest
		if err := c.BindJSON(&request); err != nil {
			c.AbortWithStatusJSON(
				http.StatusBadRequest,
				httpserver.Failure(common.NewErrorWithDescriptionValue(httpserver.CodeBadRequest, err.Error())),
			)
			return
		}
		handler.Login(
			request,
			func(status httpserver.StatusCode, response httpserver.Response[httpserver.Session]) {
				c.JSON(int(status), response)
			},
			func(status httpserver.StatusCode, response httpserver.Response[httpserver.Error]) {
				c.AbortWithStatusJSON(int(status), response)
			},
		)
	})
	router.PUT("/auth/refresh", tokenChecker.handler, func(c *gin.Context) {
		var request auth.RefreshRequest
		if err := c.BindJSON(&request); err != nil {
			c.AbortWithStatusJSON(
				http.StatusBadRequest,
				httpserver.Failure(common.NewErrorWithDescriptionValue(httpserver.CodeBadRequest, err.Error())),
			)
			return
		}
		handler.Refresh(
			request,
			func(status httpserver.StatusCode, response httpserver.Response[httpserver.Session]) {
				c.JSON(int(status), response)
			},
			func(status httpserver.StatusCode, response httpserver.Response[httpserver.Error]) {
				c.AbortWithStatusJSON(int(status), response)
			},
		)
	})
	router.PUT("/auth/updateEmail", tokenChecker.handler, func(c *gin.Context) {
		var request auth.UpdateEmailRequest
		if err := c.BindJSON(&request); err != nil {
			c.AbortWithStatusJSON(
				http.StatusBadRequest,
				httpserver.Failure(common.NewErrorWithDescriptionValue(httpserver.CodeBadRequest, err.Error())),
			)
			return
		}
		handler.UpdateEmail(
			httpserver.UserId(tokenChecker.accessToken(c)),
			request,
			func(status httpserver.StatusCode, response httpserver.Response[httpserver.Session]) {
				c.JSON(int(status), response)
			},
			func(status httpserver.StatusCode, response httpserver.Response[httpserver.Error]) {
				c.AbortWithStatusJSON(int(status), response)
			},
		)
	})
	router.PUT("/auth/updatePassword", tokenChecker.handler, func(c *gin.Context) {
		var request auth.UpdatePasswordRequest
		if err := c.BindJSON(&request); err != nil {
			c.AbortWithStatusJSON(
				http.StatusBadRequest,
				httpserver.Failure(common.NewErrorWithDescriptionValue(httpserver.CodeBadRequest, err.Error())),
			)
			return
		}
		handler.UpdatePassword(
			httpserver.UserId(tokenChecker.accessToken(c)),
			request,
			func(status httpserver.StatusCode, response httpserver.Response[httpserver.Session]) {
				c.JSON(int(status), response)
			},
			func(status httpserver.StatusCode, response httpserver.Response[httpserver.Error]) {
				c.AbortWithStatusJSON(int(status), response)
			},
		)
	})
	router.DELETE("/auth/logout", tokenChecker.handler, func(c *gin.Context) {
		handler.Logout(
			httpserver.UserId(tokenChecker.accessToken(c)),
			func(status httpserver.StatusCode, response httpserver.VoidResponse) {
				c.JSON(int(status), response)
			},
			func(status httpserver.StatusCode, response httpserver.Response[httpserver.Error]) {
				c.AbortWithStatusJSON(int(status), response)
			},
		)
	})
	router.PUT("/auth/registerForPushNotifications", tokenChecker.handler, func(c *gin.Context) {
		var request auth.RegisterForPushNotificationsRequest
		if err := c.BindJSON(&request); err != nil {
			c.AbortWithStatusJSON(
				http.StatusBadRequest,
				httpserver.Failure(common.NewErrorWithDescriptionValue(httpserver.CodeBadRequest, err.Error())),
			)
			return
		}
		handler.RegisterForPushNotifications(
			httpserver.UserId(tokenChecker.accessToken(c)),
			request,
			func(status httpserver.StatusCode, response httpserver.VoidResponse) {
				c.JSON(int(status), response)
			},
			func(status httpserver.StatusCode, response httpserver.Response[httpserver.Error]) {
				c.AbortWithStatusJSON(int(status), response)
			},
		)
	})
}

func registerRoutesSpendings(router *gin.Engine, tokenChecker ginAccessTokenChecker, handler spendings.RequestsHandler) {
	spendingsGroup := router.Group("/spendings", tokenChecker.handler)
	spendingsGroup.POST("/addExpense", func(c *gin.Context) {
		var request spendings.AddExpenseRequest
		if err := c.BindJSON(&request); err != nil {
			c.AbortWithStatusJSON(
				http.StatusBadRequest,
				httpserver.Failure(common.NewErrorWithDescriptionValue(httpserver.CodeBadRequest, err.Error())),
			)
			return
		}
		handler.AddExpense(
			httpserver.UserId(tokenChecker.accessToken(c)),
			request,
			func(status httpserver.StatusCode, response httpserver.Response[httpserver.IdentifiableExpense]) {
				c.JSON(int(status), response)
			},
			func(status httpserver.StatusCode, response httpserver.Response[httpserver.Error]) {
				c.AbortWithStatusJSON(int(status), response)
			},
		)
	})
	spendingsGroup.POST("/removeExpense", func(c *gin.Context) {
		var request spendings.RemoveExpenseRequest
		if err := c.BindJSON(&request); err != nil {
			c.AbortWithStatusJSON(
				http.StatusBadRequest,
				httpserver.Failure(common.NewErrorWithDescriptionValue(httpserver.CodeBadRequest, err.Error())),
			)
			return
		}
		handler.RemoveExpense(
			httpserver.UserId(tokenChecker.accessToken(c)),
			request,
			func(status httpserver.StatusCode, response httpserver.Response[httpserver.IdentifiableExpense]) {
				c.JSON(int(status), response)
			},
			func(status httpserver.StatusCode, response httpserver.Response[httpserver.Error]) {
				c.AbortWithStatusJSON(int(status), response)
			},
		)
	})
	spendingsGroup.GET("/getBalance", func(c *gin.Context) {
		handler.GetBalance(
			httpserver.UserId(tokenChecker.accessToken(c)),
			func(status httpserver.StatusCode, response httpserver.Response[[]httpserver.Balance]) {
				c.JSON(int(status), response)
			},
			func(status httpserver.StatusCode, response httpserver.Response[httpserver.Error]) {
				c.AbortWithStatusJSON(int(status), response)
			},
		)
	})
	spendingsGroup.GET("/getExpenses", func(c *gin.Context) {
		var request spendings.GetExpensesRequest
		if err := c.BindJSON(&request); err != nil {
			c.AbortWithStatusJSON(
				http.StatusBadRequest,
				httpserver.Failure(common.NewErrorWithDescriptionValue(httpserver.CodeBadRequest, err.Error())),
			)
			return
		}
		handler.GetExpenses(
			httpserver.UserId(tokenChecker.accessToken(c)),
			request,
			func(status httpserver.StatusCode, response httpserver.Response[[]httpserver.IdentifiableExpense]) {
				c.JSON(int(status), response)
			},
			func(status httpserver.StatusCode, response httpserver.Response[httpserver.Error]) {
				c.AbortWithStatusJSON(int(status), response)
			},
		)
	})
	spendingsGroup.GET("/getExpense", func(c *gin.Context) {
		var request spendings.GetExpenseRequest
		if err := c.BindJSON(&request); err != nil {
			c.AbortWithStatusJSON(
				http.StatusBadRequest,
				httpserver.Failure(common.NewErrorWithDescriptionValue(httpserver.CodeBadRequest, err.Error())),
			)
			return
		}
		handler.GetExpense(
			httpserver.UserId(tokenChecker.accessToken(c)),
			request,
			func(status httpserver.StatusCode, response httpserver.Response[httpserver.IdentifiableExpense]) {
				c.JSON(int(status), response)
			},
			func(status httpserver.StatusCode, response httpserver.Response[httpserver.Error]) {
				c.AbortWithStatusJSON(int(status), response)
			},
		)
	})
}

func registerRoutesFriends(router *gin.Engine, tokenChecker ginAccessTokenChecker, handler friends.RequestsHandler) {
	friendsGroup := router.Group("/friends", tokenChecker.handler)
	friendsGroup.POST("/acceptRequest", func(c *gin.Context) {
		var request friends.AcceptFriendRequest
		if err := c.BindJSON(&request); err != nil {
			c.AbortWithStatusJSON(
				http.StatusBadRequest,
				httpserver.Failure(common.NewErrorWithDescriptionValue(httpserver.CodeBadRequest, err.Error())),
			)
			return
		}
		handler.AcceptRequest(
			httpserver.UserId(tokenChecker.accessToken(c)),
			request,
			func(status httpserver.StatusCode, response httpserver.VoidResponse) {
				c.JSON(int(status), response)
			},
			func(status httpserver.StatusCode, response httpserver.Response[httpserver.Error]) {
				c.AbortWithStatusJSON(int(status), response)
			},
		)
	})
	friendsGroup.GET("/get", func(c *gin.Context) {
		var request friends.GetFriendsRequest
		if err := c.BindJSON(&request); err != nil {
			c.AbortWithStatusJSON(
				http.StatusBadRequest,
				httpserver.Failure(common.NewErrorWithDescriptionValue(httpserver.CodeBadRequest, err.Error())),
			)
			return
		}
		handler.GetFriends(
			httpserver.UserId(tokenChecker.accessToken(c)),
			request,
			func(status httpserver.StatusCode, response httpserver.Response[map[httpserver.FriendStatus][]httpserver.UserId]) {
				c.JSON(int(status), response)
			},
			func(status httpserver.StatusCode, response httpserver.Response[httpserver.Error]) {
				c.AbortWithStatusJSON(int(status), response)
			},
		)
	})
	friendsGroup.POST("/rejectRequest", func(c *gin.Context) {
		var request friends.RejectFriendRequest
		if err := c.BindJSON(&request); err != nil {
			c.AbortWithStatusJSON(
				http.StatusBadRequest,
				httpserver.Failure(common.NewErrorWithDescriptionValue(httpserver.CodeBadRequest, err.Error())),
			)
			return
		}
		handler.RejectRequest(
			httpserver.UserId(tokenChecker.accessToken(c)),
			request,
			func(status httpserver.StatusCode, response httpserver.VoidResponse) {
				c.JSON(int(status), response)
			},
			func(status httpserver.StatusCode, response httpserver.Response[httpserver.Error]) {
				c.AbortWithStatusJSON(int(status), response)
			},
		)
	})
	friendsGroup.POST("/rollbackRequest", func(c *gin.Context) {
		var request friends.RollbackFriendRequest
		if err := c.BindJSON(&request); err != nil {
			c.AbortWithStatusJSON(
				http.StatusBadRequest,
				httpserver.Failure(common.NewErrorWithDescriptionValue(httpserver.CodeBadRequest, err.Error())),
			)
			return
		}
		handler.RollbackRequest(
			httpserver.UserId(tokenChecker.accessToken(c)),
			request,
			func(status httpserver.StatusCode, response httpserver.VoidResponse) {
				c.JSON(int(status), response)
			},
			func(status httpserver.StatusCode, response httpserver.Response[httpserver.Error]) {
				c.AbortWithStatusJSON(int(status), response)
			},
		)
	})
	friendsGroup.POST("/sendRequest", func(c *gin.Context) {
		var request friends.SendFriendRequest
		if err := c.BindJSON(&request); err != nil {
			c.AbortWithStatusJSON(
				http.StatusBadRequest,
				httpserver.Failure(common.NewErrorWithDescriptionValue(httpserver.CodeBadRequest, err.Error())),
			)
			return
		}
		handler.SendRequest(
			httpserver.UserId(tokenChecker.accessToken(c)),
			request,
			func(status httpserver.StatusCode, response httpserver.VoidResponse) {
				c.JSON(int(status), response)
			},
			func(status httpserver.StatusCode, response httpserver.Response[httpserver.Error]) {
				c.AbortWithStatusJSON(int(status), response)
			},
		)
	})
	friendsGroup.POST("/unfriend", func(c *gin.Context) {
		var request friends.UnfriendRequest
		if err := c.BindJSON(&request); err != nil {
			c.AbortWithStatusJSON(
				http.StatusBadRequest,
				httpserver.Failure(common.NewErrorWithDescriptionValue(httpserver.CodeBadRequest, err.Error())),
			)
			return
		}
		handler.Unfriend(
			httpserver.UserId(tokenChecker.accessToken(c)),
			request,
			func(status httpserver.StatusCode, response httpserver.VoidResponse) {
				c.JSON(int(status), response)
			},
			func(status httpserver.StatusCode, response httpserver.Response[httpserver.Error]) {
				c.AbortWithStatusJSON(int(status), response)
			},
		)
	})
}

func registerRoutesProfile(router *gin.Engine, tokenChecker ginAccessTokenChecker, handler profile.RequestsHandler) {
	profileGroup := router.Group("/profile", tokenChecker.handler)
	profileGroup.GET("/getInfo", func(c *gin.Context) {
		handler.GetInfo(
			httpserver.UserId(tokenChecker.accessToken(c)),
			func(status httpserver.StatusCode, response httpserver.Response[httpserver.Profile]) {
				c.JSON(int(status), response)
			},
			func(status httpserver.StatusCode, response httpserver.Response[httpserver.Error]) {
				c.AbortWithStatusJSON(int(status), response)
			},
		)
	})
	profileGroup.PUT("/setAvatar", func(c *gin.Context) {
		var request profile.SetAvatarRequest
		if err := c.BindJSON(&request); err != nil {
			c.AbortWithStatusJSON(
				http.StatusBadRequest,
				httpserver.Failure(common.NewErrorWithDescriptionValue(httpserver.CodeBadRequest, err.Error())),
			)
			return
		}
		handler.SetAvatar(
			httpserver.UserId(tokenChecker.accessToken(c)),
			request,
			func(status httpserver.StatusCode, response httpserver.Response[httpserver.ImageId]) {
				c.JSON(int(status), response)
			},
			func(status httpserver.StatusCode, response httpserver.Response[httpserver.Error]) {
				c.AbortWithStatusJSON(int(status), response)
			},
		)
	})
	profileGroup.PUT("/setDisplayName", func(c *gin.Context) {
		var request profile.SetDisplayNameRequest
		if err := c.BindJSON(&request); err != nil {
			c.AbortWithStatusJSON(
				http.StatusBadRequest,
				httpserver.Failure(common.NewErrorWithDescriptionValue(httpserver.CodeBadRequest, err.Error())),
			)
			return
		}
		handler.SetDisplayName(
			httpserver.UserId(tokenChecker.accessToken(c)),
			request,
			func(status httpserver.StatusCode, response httpserver.VoidResponse) {
				c.JSON(int(status), response)
			},
			func(status httpserver.StatusCode, response httpserver.Response[httpserver.Error]) {
				c.AbortWithStatusJSON(int(status), response)
			},
		)
	})
}

func registerRoutesVerification(router *gin.Engine, tokenChecker ginAccessTokenChecker, handler verification.RequestsHandler) {
	verificationGroup := router.Group("/verification", tokenChecker.handler)
	verificationGroup.PUT("/confirmEmail", func(c *gin.Context) {
		var request verification.ConfirmEmailRequest
		if err := c.BindJSON(&request); err != nil {
			c.AbortWithStatusJSON(
				http.StatusBadRequest,
				httpserver.Failure(common.NewErrorWithDescriptionValue(httpserver.CodeBadRequest, err.Error())),
			)
			return
		}
		handler.ConfirmEmail(
			httpserver.UserId(tokenChecker.accessToken(c)),
			request,
			func(status httpserver.StatusCode, response httpserver.VoidResponse) {
				c.JSON(int(status), response)
			},
			func(status httpserver.StatusCode, response httpserver.Response[httpserver.Error]) {
				c.AbortWithStatusJSON(int(status), response)
			},
		)
	})
	verificationGroup.PUT("/sendEmailConfirmationCode", func(c *gin.Context) {
		handler.SendEmailConfirmationCode(
			httpserver.UserId(tokenChecker.accessToken(c)),
			func(status httpserver.StatusCode, response httpserver.VoidResponse) {
				c.JSON(int(status), response)
			},
			func(status httpserver.StatusCode, response httpserver.Response[httpserver.Error]) {
				c.AbortWithStatusJSON(int(status), response)
			},
		)
	})
}

func registerRoutesUsers(router *gin.Engine, tokenChecker ginAccessTokenChecker, handler users.RequestsHandler) {
	usersGroup := router.Group("/users", tokenChecker.handler)
	usersGroup.GET("/get", func(c *gin.Context) {
		var request users.GetUsersRequest
		if err := c.BindJSON(&request); err != nil {
			c.AbortWithStatusJSON(
				http.StatusBadRequest,
				httpserver.Failure(common.NewErrorWithDescriptionValue(httpserver.CodeBadRequest, err.Error())),
			)
			return
		}
		handler.GetUsers(
			httpserver.UserId(tokenChecker.accessToken(c)),
			request,
			func(status httpserver.StatusCode, response httpserver.Response[[]httpserver.User]) {
				c.JSON(int(status), response)
			},
			func(status httpserver.StatusCode, response httpserver.Response[httpserver.Error]) {
				c.AbortWithStatusJSON(int(status), response)
			},
		)
	})
}

func registerRoutesAvatars(router *gin.Engine, handler avatars.RequestsHandler) {
	avatarsGroup := router.Group("/avatars")
	avatarsGroup.GET("/avatars/get", func(c *gin.Context) {
		var request avatars.GetAvatarsRequest
		if err := c.BindJSON(&request); err != nil {
			c.AbortWithStatusJSON(
				http.StatusBadRequest,
				httpserver.Failure(common.NewErrorWithDescriptionValue(httpserver.CodeBadRequest, err.Error())),
			)
			return
		}
		handler.GetAvatars(
			request,
			func(status httpserver.StatusCode, response httpserver.Response[map[httpserver.ImageId]httpserver.Image]) {
				c.JSON(int(status), response)
			},
			func(status httpserver.StatusCode, response httpserver.Response[httpserver.Error]) {
				c.AbortWithStatusJSON(int(status), response)
			},
		)
	})
}
