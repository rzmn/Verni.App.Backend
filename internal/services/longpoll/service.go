package longpoll

import (
	authRepository "verni/internal/repositories/auth"
	"verni/internal/services/logging"

	"github.com/gin-gonic/gin"
	"github.com/jcuga/golongpoll"
)

type UserId string
type AuthRepository authRepository.Repository

type Service interface {
	CounterpartiesUpdated(uid UserId)
	ExpensesUpdated(uid UserId, counterparty UserId)
	FriendsUpdated(uid UserId)
}

func GinService(
	e *gin.Engine,
	logger logging.Service,
	accessTokenMiddleware gin.HandlerFunc,
) Service {
	op := "longpoll.GinService"
	logger.LogInfo("%s: start", op)
	longpoll, err := golongpoll.StartLongpoll(golongpoll.Options{})
	if err != nil {
		logger.LogFatal("%s: failed err: %v", op, err)
		return &ginService{}
	}
	logger.LogInfo("%s: success", op)
	e.GET("/queue/subscribe", accessTokenMiddleware, func(c *gin.Context) {
		longpoll.SubscriptionHandler(c.Writer, c.Request)
	})
	return &ginService{
		longPoll: longpoll,
		logger:   logger,
	}
}
