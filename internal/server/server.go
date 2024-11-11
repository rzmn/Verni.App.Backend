package server

import (
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

type Server interface {
	ListenAndServe()
}

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
	accessToken func(c *gin.Context) httpserver.UserId
}

const (
	accessTokenSubjectKey = "verni-subject"
)

func GinServer(
	config GinConfig,
	accessTokenChecker accessToken.RequestHandler,
	requestHandlersBuilder func(longpoll longpoll.Service) RequestHandlers,
	logger logging.Service,
) Server {
	server := createGinServer(
		config,
		accessTokenChecker,
		requestHandlersBuilder,
		logger,
	)
	return &server
}
