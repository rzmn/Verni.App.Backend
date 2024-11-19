package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/rzmn/Verni.App.Backend/internal/db"
	postgresDb "github.com/rzmn/Verni.App.Backend/internal/db/postgres"
	authRepository "github.com/rzmn/Verni.App.Backend/internal/repositories/auth"
	defaultAuthRepository "github.com/rzmn/Verni.App.Backend/internal/repositories/auth/default"
	friendsRepository "github.com/rzmn/Verni.App.Backend/internal/repositories/friends"
	defaultFriendsRepository "github.com/rzmn/Verni.App.Backend/internal/repositories/friends/default"
	imagesRepository "github.com/rzmn/Verni.App.Backend/internal/repositories/images"
	defaultImagesRepository "github.com/rzmn/Verni.App.Backend/internal/repositories/images/default"
	pushRegistryRepository "github.com/rzmn/Verni.App.Backend/internal/repositories/pushNotifications"
	defaultPushRegistryRepository "github.com/rzmn/Verni.App.Backend/internal/repositories/pushNotifications/default"
	spendingsRepository "github.com/rzmn/Verni.App.Backend/internal/repositories/spendings"
	defaultSpendingsRepository "github.com/rzmn/Verni.App.Backend/internal/repositories/spendings/default"
	usersRepository "github.com/rzmn/Verni.App.Backend/internal/repositories/users"
	defaultUsersRepository "github.com/rzmn/Verni.App.Backend/internal/repositories/users/default"
	verificationRepository "github.com/rzmn/Verni.App.Backend/internal/repositories/verification"
	defaultVerificationRepository "github.com/rzmn/Verni.App.Backend/internal/repositories/verification/default"
	ginServer "github.com/rzmn/Verni.App.Backend/internal/server/gin"

	defaultAccessTokenHandler "github.com/rzmn/Verni.App.Backend/internal/requestHandlers/accessToken/default"
	defaultAuthHandler "github.com/rzmn/Verni.App.Backend/internal/requestHandlers/auth/default"
	defaultAvatarsHandler "github.com/rzmn/Verni.App.Backend/internal/requestHandlers/avatars/default"
	defaultFriendsHandler "github.com/rzmn/Verni.App.Backend/internal/requestHandlers/friends/default"
	defaultProfileHandler "github.com/rzmn/Verni.App.Backend/internal/requestHandlers/profile/default"
	defaultSpendingsHandler "github.com/rzmn/Verni.App.Backend/internal/requestHandlers/spendings/default"
	defaultUsersHandler "github.com/rzmn/Verni.App.Backend/internal/requestHandlers/users/default"
	defaultVerificationHandler "github.com/rzmn/Verni.App.Backend/internal/requestHandlers/verification/default"

	"github.com/rzmn/Verni.App.Backend/internal/server"

	"github.com/rzmn/Verni.App.Backend/internal/services/emailSender"
	yandexEmailSender "github.com/rzmn/Verni.App.Backend/internal/services/emailSender/yandex"
	"github.com/rzmn/Verni.App.Backend/internal/services/formatValidation"
	defaultFormatValidation "github.com/rzmn/Verni.App.Backend/internal/services/formatValidation/default"
	"github.com/rzmn/Verni.App.Backend/internal/services/jwt"
	defaultJwtService "github.com/rzmn/Verni.App.Backend/internal/services/jwt/default"
	"github.com/rzmn/Verni.App.Backend/internal/services/logging"
	prodLoggingService "github.com/rzmn/Verni.App.Backend/internal/services/logging/prod"
	"github.com/rzmn/Verni.App.Backend/internal/services/pathProvider"
	envBasedPathProvider "github.com/rzmn/Verni.App.Backend/internal/services/pathProvider/env"
	"github.com/rzmn/Verni.App.Backend/internal/services/pushNotifications"
	applePushNotifications "github.com/rzmn/Verni.App.Backend/internal/services/pushNotifications/apns"
	"github.com/rzmn/Verni.App.Backend/internal/services/realtimeEvents"
	"github.com/rzmn/Verni.App.Backend/internal/services/watchdog"
	telegramWatchdog "github.com/rzmn/Verni.App.Backend/internal/services/watchdog/telegram"

	authController "github.com/rzmn/Verni.App.Backend/internal/controllers/auth"
	defaultAuthController "github.com/rzmn/Verni.App.Backend/internal/controllers/auth/default"
	avatarsController "github.com/rzmn/Verni.App.Backend/internal/controllers/avatars"
	defaultAvatarsController "github.com/rzmn/Verni.App.Backend/internal/controllers/avatars/default"
	friendsController "github.com/rzmn/Verni.App.Backend/internal/controllers/friends"
	defaultFriendsController "github.com/rzmn/Verni.App.Backend/internal/controllers/friends/default"
	profileController "github.com/rzmn/Verni.App.Backend/internal/controllers/profile"
	defaultProfileController "github.com/rzmn/Verni.App.Backend/internal/controllers/profile/default"
	spendingsController "github.com/rzmn/Verni.App.Backend/internal/controllers/spendings"
	defaultSpendingsController "github.com/rzmn/Verni.App.Backend/internal/controllers/spendings/default"
	usersController "github.com/rzmn/Verni.App.Backend/internal/controllers/users"
	defaultUsersController "github.com/rzmn/Verni.App.Backend/internal/controllers/users/default"
	verificationController "github.com/rzmn/Verni.App.Backend/internal/controllers/verification"
	defaultVerificationController "github.com/rzmn/Verni.App.Backend/internal/controllers/verification/default"
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
