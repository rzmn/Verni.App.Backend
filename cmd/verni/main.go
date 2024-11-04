package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"verni/internal/controllers/verification"
	"verni/internal/db"
	"verni/internal/http-server/handlers/auth"
	"verni/internal/http-server/handlers/avatars"
	"verni/internal/http-server/handlers/friends"
	"verni/internal/http-server/handlers/profile"
	"verni/internal/http-server/handlers/spendings"
	"verni/internal/http-server/handlers/users"
	"verni/internal/http-server/longpoll"
	"verni/internal/http-server/middleware"
	authRepository "verni/internal/repositories/auth"
	friendsRepository "verni/internal/repositories/friends"
	imagesRepository "verni/internal/repositories/images"
	pushRegistryRepository "verni/internal/repositories/pushNotifications"
	spendingsRepository "verni/internal/repositories/spendings"
	usersRepository "verni/internal/repositories/users"
	verificationRepository "verni/internal/repositories/verification"
	"verni/internal/services/emailSender"
	"verni/internal/services/formatValidation"
	"verni/internal/services/jwt"
	"verni/internal/services/logging"
	"verni/internal/services/pathProvider"
	"verni/internal/services/pushNotifications"

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

func main() {
	logger, pathProvider := func() (logging.Service, pathProvider.Service) {
		startupTime := time.Now()
		var loggingPathRef *string = nil
		logger := logging.FileLoggerService(func() *string {
			return loggingPathRef
		})
		pathProvider := pathProvider.VerniEnvService(logger)
		loggingPath := pathProvider.AbsolutePath(fmt.Sprintf("./session[%s].log", startupTime.Format("2006.01.02 15:04:05")))
		loggingPathRef = &loggingPath
		return logger, pathProvider
	}()
	configFile, err := os.Open(pathProvider.AbsolutePath("./config/prod/verni.json"))
	if err != nil {
		logger.Fatalf("failed to open config file: %s", err)
	}
	defer configFile.Close()
	configData, err := io.ReadAll(configFile)
	if err != nil {
		logger.Fatalf("failed to read config file: %s", err)
	}
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
	}
	var config Config
	json.Unmarshal([]byte(configData), &config)
	logger.Log("initializing with config %v", config)

	database := func() db.DB {
		switch config.Storage.Type {
		case "postgres":
			data, err := json.Marshal(config.Storage.Config)
			if err != nil {
				logger.Fatalf("failed to serialize ydb config err: %v", err)
			}
			var postgresConfig db.PostgresConfig
			json.Unmarshal(data, &postgresConfig)
			logger.Log("creating postgres with config %v", postgresConfig)
			db, err := db.Postgres(postgresConfig, logger)
			if err != nil {
				logger.Fatalf("failed to initialize postgres err: %v", err)
			}
			logger.Log("initialized postgres")
			return db
		default:
			logger.Fatalf("unknown storage type %s", config.Storage.Type)
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
					logger.Fatalf("failed to serialize apple apns config err: %v", err)
				}
				var apnsConfig pushNotifications.ApnsConfig
				json.Unmarshal(data, &apnsConfig)
				logger.Log("creating apple apns service with config %v", apnsConfig)
				service, err := pushNotifications.ApnsService(apnsConfig, logger, pathProvider, repositories.pushRegistry)
				if err != nil {
					logger.Fatalf("failed to initialize apple apns service err: %v", err)
				}
				logger.Log("initialized apple apns service")
				return service
			default:
				logger.Fatalf("unknown apns type %s", config.PushNotifications.Type)
				return nil
			}
		}(),
		jwt: func() jwt.Service {
			switch config.Jwt.Type {
			case "default":
				data, err := json.Marshal(config.Jwt.Config)
				if err != nil {
					logger.Fatalf("failed to serialize jwt config err: %v", err)
				}
				var defaultConfig jwt.DefaultConfig
				json.Unmarshal(data, &defaultConfig)
				logger.Log("creating jwt token service with config %v", defaultConfig)
				return jwt.DefaultService(
					defaultConfig,
					logger,
					func() time.Time {
						return time.Now()
					},
				)
			default:
				logger.Fatalf("unknown jwt service type %s", config.Jwt.Type)
				return nil
			}
		}(),
		emailSender: func() emailSender.Service {
			switch config.EmailSender.Type {
			case "yandex":
				data, err := json.Marshal(config.EmailSender.Config)
				if err != nil {
					logger.Fatalf("failed to serialize yandex email sender config err: %v", err)
				}
				var yandexConfig emailSender.YandexConfig
				json.Unmarshal(data, &yandexConfig)
				logger.Log("creating yandex email sender with config %v", yandexConfig)
				return emailSender.YandexService(yandexConfig, logger)
			default:
				logger.Fatalf("unknown email sender type %s", config.EmailSender.Type)
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
			services.push,
			logger,
		),
		users: usersController.DefaultController(
			repositories.users,
			repositories.friends,
			logger,
		),
		verification: verification.DefaultController(
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
				logger.Fatalf("failed to serialize default server config err: %v", err)
			}
			var ginConfig GinConfig
			json.Unmarshal(data, &ginConfig)
			logger.Log("creating gin server with config %v", ginConfig)
			gin.SetMode(ginConfig.RunMode)
			router := gin.New()
			tokenChecker := middleware.JwsAccessTokenCheck(
				repositories.auth,
				services.jwt,
				logger,
			)
			longpollService := longpoll.DefaultService(router, logger, tokenChecker)

			longpollService.RegisterRoutes()

			auth.RegisterRoutes(router, tokenChecker, controllers.auth)
			profile.RegisterRoutes(router, tokenChecker, controllers.profile)
			avatars.RegisterRoutes(router, controllers.avatars)
			users.RegisterRoutes(router, tokenChecker, controllers.users)
			spendings.RegisterRoutes(router, tokenChecker, controllers.spendings)
			friends.RegisterRoutes(router, tokenChecker, controllers.friends)

			address := ":" + ginConfig.Port
			return http.Server{
				Addr:         address,
				Handler:      router,
				ReadTimeout:  time.Second * time.Duration(ginConfig.IdleTimeoutSec),
				WriteTimeout: time.Second * time.Duration(ginConfig.IdleTimeoutSec),
			}
		default:
			logger.Fatalf("unknown server type %s", config.Server.Type)
			return http.Server{}
		}
	}()
	logger.Log("[info] start http server listening %s", server.Addr)
	server.ListenAndServe()
}
