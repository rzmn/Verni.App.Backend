package longpoll

import (
	"verni/internal/http-server/middleware"
	authRepository "verni/internal/repositories/auth"
	"verni/internal/services/logging"

	"github.com/gin-gonic/gin"
)

type UserId string
type AuthRepository authRepository.Repository

type Service interface {
	CounterpartiesUpdated(uid UserId)
	ExpensesUpdated(uid UserId, counterparty UserId)
	FriendsUpdated(uid UserId)
	RegisterRoutes()
}

func DefaultService(e *gin.Engine, logger logging.Service, tokenChecker middleware.AccessTokenChecker) Service {
	return &defaultService{
		engine:       e,
		tokenChecker: tokenChecker,
		logger:       logger,
	}
}
