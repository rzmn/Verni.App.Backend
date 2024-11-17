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
	defaultAuthRepository "verni/internal/repositories/auth/default"
	friendsRepository "verni/internal/repositories/friends"
	defaultFriendsRepository "verni/internal/repositories/friends/default"
	imagesRepository "verni/internal/repositories/images"
	defaultImagesRepository "verni/internal/repositories/images/default"
	pushRegistryRepository "verni/internal/repositories/pushNotifications"
	defaultPushRegistryRepository "verni/internal/repositories/pushNotifications/default"
	spendingsRepository "verni/internal/repositories/spendings"
	defaultSpendingsRepository "verni/internal/repositories/spendings/default"
	usersRepository "verni/internal/repositories/users"
	defaultUsersRepository "verni/internal/repositories/users/default"
	verificationRepository "verni/internal/repositories/verification"
	defaultVerificationRepository "verni/internal/repositories/verification/default"
	ginServer "verni/internal/server/gin"

	defaultAccessTokenHandler "verni/internal/requestHandlers/accessToken/default"
	defaultAuthHandler "verni/internal/requestHandlers/auth/default"
	defaultAvatarsHandler "verni/internal/requestHandlers/avatars/default"
	defaultFriendsHandler "verni/internal/requestHandlers/friends/default"
	defaultProfileHandler "verni/internal/requestHandlers/profile/default"
	defaultSpendingsHandler "verni/internal/requestHandlers/spendings/default"
	defaultUsersHandler "verni/internal/requestHandlers/users/default"
	defaultVerificationHandler "verni/internal/requestHandlers/verification/default"

	"verni/internal/server"

	"verni/internal/services/emailSender"
	yandexEmailSender "verni/internal/services/emailSender/yandex"
	"verni/internal/services/formatValidation"
	defaultFormatValidation "verni/internal/services/formatValidation/default"
	"verni/internal/services/jwt"
	defaultJwtService "verni/internal/services/jwt/default"
	"verni/internal/services/logging"
	prodLoggingService "verni/internal/services/logging/prod"
	"verni/internal/services/pathProvider"
	envBasedPathProvider "verni/internal/services/pathProvider/env"
	"verni/internal/services/pushNotifications"
	applePushNotifications "verni/internal/services/pushNotifications/apns"
	"verni/internal/services/realtimeEvents"
	"verni/internal/services/watchdog"
	telegramWatchdog "verni/internal/services/watchdog/telegram"

	authController "verni/internal/controllers/auth"
	defaultAuthController "verni/internal/controllers/auth/default"
	avatarsController "verni/internal/controllers/avatars"
	defaultAvatarsController "verni/internal/controllers/avatars/default"
	friendsController "verni/internal/controllers/friends"
	defaultFriendsController "verni/internal/controllers/friends/default"
	profileController "verni/internal/controllers/profile"
	defaultProfileController "verni/internal/controllers/profile/default"
	spendingsController "verni/internal/controllers/spendings"
	defaultSpendingsController "verni/internal/controllers/spendings/default"
	usersController "verni/internal/controllers/users"
	defaultUsersController "verni/internal/controllers/users/default"
	verificationController "verni/internal/controllers/verification"
	defaultVerificationController "verni/internal/controllers/verification/default"
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
		logger := prodLoggingService.New(func() *prodLoggingService.ProdLoggerConfig {
			if loggingDirectoryRef == nil {
				return nil
			}
			if watchdogRef == nil {
				return nil
			}
			return &prodLoggingService.ProdLoggerConfig{
				Watchdog:         *watchdogRef,
				LoggingDirectory: *loggingDirectoryRef,
			}
		})
		pathProvider := envBasedPathProvider.New(logger)
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
				var telegramConfig telegramWatchdog.TelegramConfig
				json.Unmarshal(data, &telegramConfig)
				logger.LogInfo("creating telegram watchdog with config %v", telegramConfig)
				watchdog, err := telegramWatchdog.New(telegramConfig)
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
		auth:         defaultAuthRepository.New(database, logger),
		friends:      defaultFriendsRepository.New(database, logger),
		images:       defaultImagesRepository.New(database, logger),
		pushRegistry: defaultPushRegistryRepository.New(database, logger),
		spendings:    defaultSpendingsRepository.New(database, logger),
		users:        defaultUsersRepository.New(database, logger),
		verification: defaultVerificationRepository.New(database, logger),
	}
	services := Services{
		push: func() pushNotifications.Service {
			switch config.PushNotifications.Type {
			case "apns":
				data, err := json.Marshal(config.PushNotifications.Config)
				if err != nil {
					logger.LogFatal("failed to serialize apple apns config err: %v", err)
				}
				var apnsConfig applePushNotifications.ApnsConfig
				json.Unmarshal(data, &apnsConfig)
				logger.LogInfo("creating apple apns service with config %v", apnsConfig)
				service, err := applePushNotifications.New(apnsConfig, logger, pathProvider, repositories.pushRegistry)
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
				var defaultConfig defaultJwtService.DefaultConfig
				json.Unmarshal(data, &defaultConfig)
				logger.LogInfo("creating jwt token service with config %v", defaultConfig)
				return defaultJwtService.New(
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
				var yandexConfig yandexEmailSender.YandexConfig
				json.Unmarshal(data, &yandexConfig)
				logger.LogInfo("creating yandex email sender with config %v", yandexConfig)
				return yandexEmailSender.New(yandexConfig, logger)
			default:
				logger.LogFatal("unknown email sender type %s", config.EmailSender.Type)
				return nil
			}
		}(),
		formatValidationService: func() formatValidation.Service {
			return defaultFormatValidation.New(logger)
		}(),
	}
	controllers := Controllers{
		auth: defaultAuthController.New(
			repositories.auth,
			repositories.pushRegistry,
			repositories.users,
			services.jwt,
			services.formatValidationService,
			logger,
		),
		avatars: defaultAvatarsController.New(
			repositories.images,
			logger,
		),
		friends: defaultFriendsController.New(
			repositories.friends,
			logger,
		),
		profile: defaultProfileController.New(
			repositories.auth,
			repositories.images,
			repositories.users,
			repositories.friends,
			services.formatValidationService,
			logger,
		),
		spendings: defaultSpendingsController.New(
			repositories.spendings,
			logger,
		),
		users: defaultUsersController.New(
			repositories.users,
			repositories.friends,
			logger,
		),
		verification: defaultVerificationController.New(
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
			var ginConfig ginServer.GinConfig
			json.Unmarshal(data, &ginConfig)
			logger.LogInfo("creating gin server with config %v", ginConfig)
			return ginServer.New(
				ginConfig,
				defaultAccessTokenHandler.New(
					repositories.auth,
					services.jwt,
					logger,
				),
				func(realtimeEvents realtimeEvents.Service) ginServer.RequestHandlers {
					return ginServer.RequestHandlers{
						Auth: defaultAuthHandler.New(
							controllers.auth,
							logger,
						),
						Spendings: defaultSpendingsHandler.New(
							controllers.spendings,
							services.push,
							realtimeEvents,
							logger,
						),
						Friends: defaultFriendsHandler.New(
							controllers.friends,
							services.push,
							realtimeEvents,
							logger,
						),
						Profile: defaultProfileHandler.New(
							controllers.profile,
							logger,
						),
						Verification: defaultVerificationHandler.New(
							controllers.verification,
							logger,
						),
						Users: defaultUsersHandler.New(
							controllers.users,
							logger,
						),
						Avatars: defaultAvatarsHandler.New(
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
