package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"verni/internal/db"
	postgresDb "verni/internal/db/postgres"
	authRepository "verni/internal/repositories/auth"
	friendsRepository "verni/internal/repositories/friends"
	imagesRepository "verni/internal/repositories/images"
	pushRegistryRepository "verni/internal/repositories/pushNotifications"
	spendingsRepository "verni/internal/repositories/spendings"
	usersRepository "verni/internal/repositories/users"
	verificationRepository "verni/internal/repositories/verification"
	"verni/internal/requestHandlers/accessToken"
	"verni/internal/requestHandlers/auth"
	"verni/internal/requestHandlers/avatars"
	"verni/internal/requestHandlers/friends"
	"verni/internal/requestHandlers/profile"
	"verni/internal/requestHandlers/spendings"
	"verni/internal/requestHandlers/users"
	"verni/internal/requestHandlers/verification"
	"verni/internal/server"
	"verni/internal/services/emailSender"
	"verni/internal/services/formatValidation"
	"verni/internal/services/jwt"
	"verni/internal/services/logging"
	"verni/internal/services/longpoll"
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
			var postgresConfig postgresDb.PostgresConfig
			json.Unmarshal(data, &postgresConfig)
			logger.LogInfo("creating postgres with config %v", postgresConfig)
			db, err := postgresDb.Postgres(postgresConfig, logger)
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
	server := func() server.Server {
		switch config.Server.Type {
		case "gin":
			data, err := json.Marshal(config.Server.Config)
			if err != nil {
				logger.LogFatal("failed to serialize default server config err: %v", err)
			}
			var ginConfig server.GinConfig
			json.Unmarshal(data, &ginConfig)
			logger.LogInfo("creating gin server with config %v", ginConfig)
			return server.GinServer(
				ginConfig,
				accessToken.DefaultHandler(
					repositories.auth,
					services.jwt,
					logger,
				),
				func(longpoll longpoll.Service) server.RequestHandlers {
					return server.RequestHandlers{
						Auth: auth.DefaultHandler(
							controllers.auth,
							logger,
						),
						Spendings: spendings.DefaultHandler(
							controllers.spendings,
							services.push,
							longpoll,
							logger,
						),
						Friends: friends.DefaultHandler(
							controllers.friends,
							services.push,
							longpoll,
							logger,
						),
						Profile: profile.DefaultHandler(
							controllers.profile,
							logger,
						),
						Verification: verification.DefaultHandler(
							controllers.verification,
							logger,
						),
						Users: users.DefaultHandler(
							controllers.users,
							logger,
						),
						Avatars: avatars.DefaultHandler(
							controllers.avatars,
							logger,
						),
					}
				},
				logger,
			)
		default:
			logger.LogFatal("unknown server type %s", config.Server.Type)
			return nil
		}
	}()
	server.ListenAndServe()
}
