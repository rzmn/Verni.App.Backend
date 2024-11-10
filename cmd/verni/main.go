package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"verni/internal/common"
	"verni/internal/db"
	httpserver "verni/internal/http-server"
	"verni/internal/http-server/handlers/avatars"
	"verni/internal/http-server/handlers/profile"
	"verni/internal/http-server/handlers/users"
	"verni/internal/http-server/longpoll"
	"verni/internal/http-server/middleware"
	"verni/internal/http-server/responses"
	authRepository "verni/internal/repositories/auth"
	friendsRepository "verni/internal/repositories/friends"
	imagesRepository "verni/internal/repositories/images"
	pushRegistryRepository "verni/internal/repositories/pushNotifications"
	spendingsRepository "verni/internal/repositories/spendings"
	usersRepository "verni/internal/repositories/users"
	verificationRepository "verni/internal/repositories/verification"
	"verni/internal/requestHandlers/auth"
	"verni/internal/requestHandlers/friends"
	"verni/internal/requestHandlers/spendings"
	"verni/internal/services/emailSender"
	"verni/internal/services/formatValidation"
	"verni/internal/services/jwt"
	"verni/internal/services/logging"
	"verni/internal/services/pathProvider"
	"verni/internal/services/pushNotifications"
	"verni/internal/services/watchdog"

	authController "verni/internal/controllers/auth"
	avatarsController "verni/internal/controllers/avatars"
	friendsController "verni/internal/controllers/friends"
	profileController "verni/internal/controllers/profile"
	spendingsController "verni/internal/controllers/spendings"
	usersController "verni/internal/controllers/users"
	verificationController "verni/internal/controllers/verification"

	"github.com/gin-gonic/gin"
)

type Repositories struct {
	auth         authRepository.Repository
	friends      friendsRepository.Repository
	images       imagesRepository.Repository
	pushRegistry pushRegistryRepository.Repository
	spendings    spendingsRepository.Repository
	users        usersRepository.Repository
	verification verificationRepository.Repository
}

type Services struct {
	push                    pushNotifications.Service
	jwt                     jwt.Service
	emailSender             emailSender.Service
	formatValidationService formatValidation.Service
}

type Controllers struct {
	auth         authController.Controller
	avatars      avatarsController.Controller
	friends      friendsController.Controller
	profile      profileController.Controller
	spendings    spendingsController.Controller
	users        usersController.Controller
	verification verificationController.Controller
}

type RequestHandlers struct {
	spendings spendings.RequestsHandler
	friends   friends.RequestsHandler
	auth      auth.RequestsHandler
}

func main() {
	type Module struct {
		Type   string                 `json:"type"`
		Config map[string]interface{} `json:"config"`
	}
	type Config struct {
		Storage           Module `json:"storage"`
		PushNotifications Module `json:"pushNotifications"`
		EmailSender       Module `json:"emailSender"`
		Jwt               Module `json:"jwt"`
		Server            Module `json:"server"`
		Watchdog          Module `json:"watchdog"`
	}
	logger, pathProvider, config := func() (logging.Service, pathProvider.Service, Config) {
		startupTime := time.Now()
		var loggingDirectoryRef *string = nil
		var watchdogRef *watchdog.Service = nil
		logger := logging.Prod(func() *logging.ProdLoggerConfig {
			if loggingDirectoryRef == nil {
				return nil
			}
			if watchdogRef == nil {
				return nil
			}
			return &logging.ProdLoggerConfig{
				Watchdog:         *watchdogRef,
				LoggingDirectory: *loggingDirectoryRef,
			}
		})
		pathProvider := pathProvider.VerniEnvService(logger)
		loggingDirectory := pathProvider.AbsolutePath(fmt.Sprintf("./session[%s].log", startupTime.Format("2006.01.02 15:04:05")))
		if err := os.MkdirAll(loggingDirectory, os.ModePerm); err != nil {
			loggingDirectoryRef = nil
		} else {
			loggingDirectoryRef = &loggingDirectory
		}
		configFile, err := os.Open(pathProvider.AbsolutePath("./config/prod/verni.json"))
		if err != nil {
			logger.LogFatal("failed to open config file: %s", err)
		}
		defer configFile.Close()
		configData, err := io.ReadAll(configFile)
		if err != nil {
			logger.LogFatal("failed to read config file: %s", err)
		}
		var config Config
		json.Unmarshal([]byte(configData), &config)
		watchdog := func() watchdog.Service {
			switch config.Watchdog.Type {
			case "telegram":
				data, err := json.Marshal(config.Watchdog.Config)
				if err != nil {
					logger.LogFatal("failed to serialize telegram watchdog config err: %v", err)
				}
				var telegramConfig watchdog.TelegramConfig
				json.Unmarshal(data, &telegramConfig)
				logger.LogInfo("creating telegram watchdog with config %v", telegramConfig)
				watchdog, err := watchdog.TelegramService(telegramConfig)
				if err != nil {
					logger.LogFatal("failed to initialize telegram watchdog err: %v", err)
				}
				logger.LogInfo("initialized postgres")
				return watchdog
			default:
				logger.LogFatal("unknown storage type %s", config.Storage.Type)
				return nil
			}
		}()
		watchdogRef = &watchdog
		return logger, pathProvider, config
	}()
	logger.LogInfo("initializing with config %v", config)

	database := func() db.DB {
		switch config.Storage.Type {
		case "postgres":
			data, err := json.Marshal(config.Storage.Config)
			if err != nil {
				logger.LogFatal("failed to serialize ydb config err: %v", err)
			}
			var postgresConfig db.PostgresConfig
			json.Unmarshal(data, &postgresConfig)
			logger.LogInfo("creating postgres with config %v", postgresConfig)
			db, err := db.Postgres(postgresConfig, logger)
			if err != nil {
				logger.LogFatal("failed to initialize postgres err: %v", err)
			}
			logger.LogInfo("initialized postgres")
			return db
		default:
			logger.LogFatal("unknown storage type %s", config.Storage.Type)
			return nil
		}
	}()
	defer database.Close()
	repositories := Repositories{
		auth:         authRepository.PostgresRepository(database, logger),
		friends:      friendsRepository.PostgresRepository(database, logger),
		images:       imagesRepository.PostgresRepository(database, logger),
		pushRegistry: pushRegistryRepository.PostgresRepository(database, logger),
		spendings:    spendingsRepository.PostgresRepository(database, logger),
		users:        usersRepository.PostgresRepository(database, logger),
		verification: verificationRepository.PostgresRepository(database, logger),
	}
	services := Services{
		push: func() pushNotifications.Service {
			switch config.PushNotifications.Type {
			case "apns":
				data, err := json.Marshal(config.PushNotifications.Config)
				if err != nil {
					logger.LogFatal("failed to serialize apple apns config err: %v", err)
				}
				var apnsConfig pushNotifications.ApnsConfig
				json.Unmarshal(data, &apnsConfig)
				logger.LogInfo("creating apple apns service with config %v", apnsConfig)
				service, err := pushNotifications.ApnsService(apnsConfig, logger, pathProvider, repositories.pushRegistry)
				if err != nil {
					logger.LogFatal("failed to initialize apple apns service err: %v", err)
				}
				logger.LogInfo("initialized apple apns service")
				return service
			default:
				logger.LogFatal("unknown apns type %s", config.PushNotifications.Type)
				return nil
			}
		}(),
		jwt: func() jwt.Service {
			switch config.Jwt.Type {
			case "default":
				data, err := json.Marshal(config.Jwt.Config)
				if err != nil {
					logger.LogFatal("failed to serialize jwt config err: %v", err)
				}
				var defaultConfig jwt.DefaultConfig
				json.Unmarshal(data, &defaultConfig)
				logger.LogInfo("creating jwt token service with config %v", defaultConfig)
				return jwt.DefaultService(
					defaultConfig,
					logger,
					func() time.Time {
						return time.Now()
					},
				)
			default:
				logger.LogFatal("unknown jwt service type %s", config.Jwt.Type)
				return nil
			}
		}(),
		emailSender: func() emailSender.Service {
			switch config.EmailSender.Type {
			case "yandex":
				data, err := json.Marshal(config.EmailSender.Config)
				if err != nil {
					logger.LogFatal("failed to serialize yandex email sender config err: %v", err)
				}
				var yandexConfig emailSender.YandexConfig
				json.Unmarshal(data, &yandexConfig)
				logger.LogInfo("creating yandex email sender with config %v", yandexConfig)
				return emailSender.YandexService(yandexConfig, logger)
			default:
				logger.LogFatal("unknown email sender type %s", config.EmailSender.Type)
				return nil
			}
		}(),
		formatValidationService: func() formatValidation.Service {
			return formatValidation.DefaultService(logger)
		}(),
	}
	controllers := Controllers{
		auth: authController.DefaultController(
			repositories.auth,
			repositories.pushRegistry,
			repositories.users,
			services.jwt,
			services.formatValidationService,
			logger,
		),
		avatars: avatarsController.DefaultController(
			repositories.images,
			logger,
		),
		friends: friendsController.DefaultController(
			repositories.friends,
			logger,
		),
		profile: profileController.DefaultController(
			repositories.auth,
			repositories.images,
			repositories.users,
			repositories.friends,
			services.formatValidationService,
			logger,
		),
		spendings: spendingsController.DefaultController(
			repositories.spendings,
			logger,
		),
		users: usersController.DefaultController(
			repositories.users,
			repositories.friends,
			logger,
		),
		verification: verificationController.DefaultController(
			repositories.verification,
			repositories.auth,
			services.emailSender,
			logger,
		),
	}
	server := func() http.Server {
		switch config.Server.Type {
		case "gin":
			type GinConfig struct {
				TimeoutSec     int    `json:"timeoutSec"`
				IdleTimeoutSec int    `json:"idleTimeoutSec"`
				RunMode        string `json:"runMode"`
				Port           string `json:"port"`
			}
			data, err := json.Marshal(config.Server.Config)
			if err != nil {
				logger.LogFatal("failed to serialize default server config err: %v", err)
			}
			var ginConfig GinConfig
			json.Unmarshal(data, &ginConfig)
			logger.LogInfo("creating gin server with config %v", ginConfig)
			gin.SetMode(ginConfig.RunMode)
			router := gin.New()
			tokenChecker := middleware.JwsAccessTokenCheck(
				repositories.auth,
				services.jwt,
				logger,
			)
			longpollService := longpoll.DefaultService(router, logger, tokenChecker)

			longpollService.RegisterRoutes()

			requestHandlers := RequestHandlers{
				spendings: spendings.DefaultHandler(
					controllers.spendings,
					services.push,
					longpollService,
					logger,
				),
				friends: friends.DefaultHandler(
					controllers.friends,
					services.push,
					longpollService,
					logger,
				),
				auth: auth.DefaultHandler(
					controllers.auth,
					logger,
				),
			}
			spendingsGroup := router.Group("/spendings", tokenChecker.Handler)
			spendingsGroup.POST("/addExpense", func(c *gin.Context) {
				var request spendings.AddExpenseRequest
				if err := c.BindJSON(&request); err != nil {
					c.AbortWithStatusJSON(
						http.StatusBadRequest,
						responses.Failure(common.NewErrorWithDescriptionValue(responses.CodeBadRequest, err.Error())),
					)
					return
				}
				requestHandlers.spendings.AddExpense(
					httpserver.UserId(tokenChecker.AccessToken(c)),
					request,
					func(status httpserver.StatusCode, response responses.Response[httpserver.IdentifiableExpense]) {
						c.JSON(int(status), response)
					},
					func(status httpserver.StatusCode, response responses.Response[responses.Error]) {
						c.AbortWithStatusJSON(int(status), response)
					},
				)
			})
			spendingsGroup.POST("/removeExpense", func(c *gin.Context) {
				var request spendings.RemoveExpenseRequest
				if err := c.BindJSON(&request); err != nil {
					c.AbortWithStatusJSON(
						http.StatusBadRequest,
						responses.Failure(common.NewErrorWithDescriptionValue(responses.CodeBadRequest, err.Error())),
					)
					return
				}
				requestHandlers.spendings.RemoveExpense(
					httpserver.UserId(tokenChecker.AccessToken(c)),
					request,
					func(status httpserver.StatusCode, response responses.Response[httpserver.IdentifiableExpense]) {
						c.JSON(int(status), response)
					},
					func(status httpserver.StatusCode, response responses.Response[responses.Error]) {
						c.AbortWithStatusJSON(int(status), response)
					},
				)
			})
			spendingsGroup.GET("/getBalance", func(c *gin.Context) {
				requestHandlers.spendings.GetBalance(
					httpserver.UserId(tokenChecker.AccessToken(c)),
					func(status httpserver.StatusCode, response responses.Response[[]httpserver.Balance]) {
						c.JSON(int(status), response)
					},
					func(status httpserver.StatusCode, response responses.Response[responses.Error]) {
						c.AbortWithStatusJSON(int(status), response)
					},
				)
			})
			spendingsGroup.GET("/getExpenses", func(c *gin.Context) {
				var request spendings.GetExpensesRequest
				if err := c.BindJSON(&request); err != nil {
					c.AbortWithStatusJSON(
						http.StatusBadRequest,
						responses.Failure(common.NewErrorWithDescriptionValue(responses.CodeBadRequest, err.Error())),
					)
					return
				}
				requestHandlers.spendings.GetExpenses(
					httpserver.UserId(tokenChecker.AccessToken(c)),
					request,
					func(status httpserver.StatusCode, response responses.Response[[]httpserver.IdentifiableExpense]) {
						c.JSON(int(status), response)
					},
					func(status httpserver.StatusCode, response responses.Response[responses.Error]) {
						c.AbortWithStatusJSON(int(status), response)
					},
				)
			})
			spendingsGroup.GET("/getExpense", func(c *gin.Context) {
				var request spendings.GetExpenseRequest
				if err := c.BindJSON(&request); err != nil {
					c.AbortWithStatusJSON(
						http.StatusBadRequest,
						responses.Failure(common.NewErrorWithDescriptionValue(responses.CodeBadRequest, err.Error())),
					)
					return
				}
				requestHandlers.spendings.GetExpense(
					httpserver.UserId(tokenChecker.AccessToken(c)),
					request,
					func(status httpserver.StatusCode, response responses.Response[httpserver.IdentifiableExpense]) {
						c.JSON(int(status), response)
					},
					func(status httpserver.StatusCode, response responses.Response[responses.Error]) {
						c.AbortWithStatusJSON(int(status), response)
					},
				)
			})
			friendsGroup := router.Group("/friends", tokenChecker.Handler)
			friendsGroup.POST("/acceptRequest", func(c *gin.Context) {
				var request friends.AcceptFriendRequest
				if err := c.BindJSON(&request); err != nil {
					c.AbortWithStatusJSON(
						http.StatusBadRequest,
						responses.Failure(common.NewErrorWithDescriptionValue(responses.CodeBadRequest, err.Error())),
					)
					return
				}
				requestHandlers.friends.AcceptRequest(
					httpserver.UserId(tokenChecker.AccessToken(c)),
					request,
					func(status httpserver.StatusCode, response responses.VoidResponse) {
						c.JSON(int(status), response)
					},
					func(status httpserver.StatusCode, response responses.Response[responses.Error]) {
						c.AbortWithStatusJSON(int(status), response)
					},
				)
			})
			friendsGroup.GET("/get", func(c *gin.Context) {
				var request friends.GetFriendsRequest
				if err := c.BindJSON(&request); err != nil {
					c.AbortWithStatusJSON(
						http.StatusBadRequest,
						responses.Failure(common.NewErrorWithDescriptionValue(responses.CodeBadRequest, err.Error())),
					)
					return
				}
				requestHandlers.friends.GetFriends(
					httpserver.UserId(tokenChecker.AccessToken(c)),
					request,
					func(status httpserver.StatusCode, response responses.Response[map[httpserver.FriendStatus][]httpserver.UserId]) {
						c.JSON(int(status), response)
					},
					func(status httpserver.StatusCode, response responses.Response[responses.Error]) {
						c.AbortWithStatusJSON(int(status), response)
					},
				)
			})
			friendsGroup.POST("/rejectRequest", func(c *gin.Context) {
				var request friends.RejectFriendRequest
				if err := c.BindJSON(&request); err != nil {
					c.AbortWithStatusJSON(
						http.StatusBadRequest,
						responses.Failure(common.NewErrorWithDescriptionValue(responses.CodeBadRequest, err.Error())),
					)
					return
				}
				requestHandlers.friends.RejectRequest(
					httpserver.UserId(tokenChecker.AccessToken(c)),
					request,
					func(status httpserver.StatusCode, response responses.VoidResponse) {
						c.JSON(int(status), response)
					},
					func(status httpserver.StatusCode, response responses.Response[responses.Error]) {
						c.AbortWithStatusJSON(int(status), response)
					},
				)
			})
			friendsGroup.POST("/rollbackRequest", func(c *gin.Context) {
				var request friends.RollbackFriendRequest
				if err := c.BindJSON(&request); err != nil {
					c.AbortWithStatusJSON(
						http.StatusBadRequest,
						responses.Failure(common.NewErrorWithDescriptionValue(responses.CodeBadRequest, err.Error())),
					)
					return
				}
				requestHandlers.friends.RollbackRequest(
					httpserver.UserId(tokenChecker.AccessToken(c)),
					request,
					func(status httpserver.StatusCode, response responses.VoidResponse) {
						c.JSON(int(status), response)
					},
					func(status httpserver.StatusCode, response responses.Response[responses.Error]) {
						c.AbortWithStatusJSON(int(status), response)
					},
				)
			})
			friendsGroup.POST("/sendRequest", func(c *gin.Context) {
				var request friends.SendFriendRequest
				if err := c.BindJSON(&request); err != nil {
					c.AbortWithStatusJSON(
						http.StatusBadRequest,
						responses.Failure(common.NewErrorWithDescriptionValue(responses.CodeBadRequest, err.Error())),
					)
					return
				}
				requestHandlers.friends.SendRequest(
					httpserver.UserId(tokenChecker.AccessToken(c)),
					request,
					func(status httpserver.StatusCode, response responses.VoidResponse) {
						c.JSON(int(status), response)
					},
					func(status httpserver.StatusCode, response responses.Response[responses.Error]) {
						c.AbortWithStatusJSON(int(status), response)
					},
				)
			})
			friendsGroup.POST("/unfriend", func(c *gin.Context) {
				var request friends.UnfriendRequest
				if err := c.BindJSON(&request); err != nil {
					c.AbortWithStatusJSON(
						http.StatusBadRequest,
						responses.Failure(common.NewErrorWithDescriptionValue(responses.CodeBadRequest, err.Error())),
					)
					return
				}
				requestHandlers.friends.Unfriend(
					httpserver.UserId(tokenChecker.AccessToken(c)),
					request,
					func(status httpserver.StatusCode, response responses.VoidResponse) {
						c.JSON(int(status), response)
					},
					func(status httpserver.StatusCode, response responses.Response[responses.Error]) {
						c.AbortWithStatusJSON(int(status), response)
					},
				)
			})
			router.PUT("/auth/signup", func(c *gin.Context) {
				var request auth.SignupRequest
				if err := c.BindJSON(&request); err != nil {
					c.AbortWithStatusJSON(
						http.StatusBadRequest,
						responses.Failure(common.NewErrorWithDescriptionValue(responses.CodeBadRequest, err.Error())),
					)
					return
				}
				requestHandlers.auth.Signup(
					request,
					func(status httpserver.StatusCode, response responses.Response[httpserver.Session]) {
						c.JSON(int(status), response)
					},
					func(status httpserver.StatusCode, response responses.Response[responses.Error]) {
						c.AbortWithStatusJSON(int(status), response)
					},
				)
			})
			router.PUT("/auth/login", func(c *gin.Context) {
				var request auth.LoginRequest
				if err := c.BindJSON(&request); err != nil {
					c.AbortWithStatusJSON(
						http.StatusBadRequest,
						responses.Failure(common.NewErrorWithDescriptionValue(responses.CodeBadRequest, err.Error())),
					)
					return
				}
				requestHandlers.auth.Login(
					request,
					func(status httpserver.StatusCode, response responses.Response[httpserver.Session]) {
						c.JSON(int(status), response)
					},
					func(status httpserver.StatusCode, response responses.Response[responses.Error]) {
						c.AbortWithStatusJSON(int(status), response)
					},
				)
			})
			router.PUT("/auth/refresh", tokenChecker.Handler, func(c *gin.Context) {
				var request auth.RefreshRequest
				if err := c.BindJSON(&request); err != nil {
					c.AbortWithStatusJSON(
						http.StatusBadRequest,
						responses.Failure(common.NewErrorWithDescriptionValue(responses.CodeBadRequest, err.Error())),
					)
					return
				}
				requestHandlers.auth.Refresh(
					request,
					func(status httpserver.StatusCode, response responses.Response[httpserver.Session]) {
						c.JSON(int(status), response)
					},
					func(status httpserver.StatusCode, response responses.Response[responses.Error]) {
						c.AbortWithStatusJSON(int(status), response)
					},
				)
			})
			router.PUT("/auth/updateEmail", tokenChecker.Handler, func(c *gin.Context) {
				var request auth.UpdateEmailRequest
				if err := c.BindJSON(&request); err != nil {
					c.AbortWithStatusJSON(
						http.StatusBadRequest,
						responses.Failure(common.NewErrorWithDescriptionValue(responses.CodeBadRequest, err.Error())),
					)
					return
				}
				requestHandlers.auth.UpdateEmail(
					httpserver.UserId(tokenChecker.AccessToken(c)),
					request,
					func(status httpserver.StatusCode, response responses.Response[httpserver.Session]) {
						c.JSON(int(status), response)
					},
					func(status httpserver.StatusCode, response responses.Response[responses.Error]) {
						c.AbortWithStatusJSON(int(status), response)
					},
				)
			})
			router.PUT("/auth/updatePassword", tokenChecker.Handler, func(c *gin.Context) {
				var request auth.UpdatePasswordRequest
				if err := c.BindJSON(&request); err != nil {
					c.AbortWithStatusJSON(
						http.StatusBadRequest,
						responses.Failure(common.NewErrorWithDescriptionValue(responses.CodeBadRequest, err.Error())),
					)
					return
				}
				requestHandlers.auth.UpdatePassword(
					httpserver.UserId(tokenChecker.AccessToken(c)),
					request,
					func(status httpserver.StatusCode, response responses.Response[httpserver.Session]) {
						c.JSON(int(status), response)
					},
					func(status httpserver.StatusCode, response responses.Response[responses.Error]) {
						c.AbortWithStatusJSON(int(status), response)
					},
				)
			})
			router.DELETE("/auth/logout", tokenChecker.Handler, func(c *gin.Context) {
				requestHandlers.auth.Logout(
					httpserver.UserId(tokenChecker.AccessToken(c)),
					func(status httpserver.StatusCode, response responses.VoidResponse) {
						c.JSON(int(status), response)
					},
					func(status httpserver.StatusCode, response responses.Response[responses.Error]) {
						c.AbortWithStatusJSON(int(status), response)
					},
				)
			})
			router.PUT("/auth/registerForPushNotifications", tokenChecker.Handler, func(c *gin.Context) {
				var request auth.RegisterForPushNotificationsRequest
				if err := c.BindJSON(&request); err != nil {
					c.AbortWithStatusJSON(
						http.StatusBadRequest,
						responses.Failure(common.NewErrorWithDescriptionValue(responses.CodeBadRequest, err.Error())),
					)
					return
				}
				requestHandlers.auth.RegisterForPushNotifications(
					httpserver.UserId(tokenChecker.AccessToken(c)),
					request,
					func(status httpserver.StatusCode, response responses.VoidResponse) {
						c.JSON(int(status), response)
					},
					func(status httpserver.StatusCode, response responses.Response[responses.Error]) {
						c.AbortWithStatusJSON(int(status), response)
					},
				)
			})
			profile.RegisterRoutes(router, logger, tokenChecker, controllers.profile)
			avatars.RegisterRoutes(router, logger, controllers.avatars)
			users.RegisterRoutes(router, logger, tokenChecker, controllers.users)

			address := ":" + ginConfig.Port
			return http.Server{
				Addr:         address,
				Handler:      router,
				ReadTimeout:  time.Second * time.Duration(ginConfig.IdleTimeoutSec),
				WriteTimeout: time.Second * time.Duration(ginConfig.IdleTimeoutSec),
			}
		default:
			logger.LogFatal("unknown server type %s", config.Server.Type)
			return http.Server{}
		}
	}()
	logger.LogInfo("[info] start http server listening %s", server.Addr)
	server.ListenAndServe()
}
