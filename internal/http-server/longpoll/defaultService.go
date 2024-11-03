package longpoll

import (
	"fmt"
	"verni/internal/http-server/middleware"
	"verni/internal/services/logging"

	"github.com/jcuga/golongpoll"

	"github.com/gin-gonic/gin"
)

type defaultService struct {
	engine       *gin.Engine
	tokenChecker middleware.AccessTokenChecker
	longPoll     *golongpoll.LongpollManager
	logger       logging.Service
}

func (s *defaultService) RegisterRoutes() {
	const op = "longpoll.defaultService.RegisterRoutes"
	s.logger.Log("%s: start", op)
	longpoll, err := golongpoll.StartLongpoll(golongpoll.Options{})
	if err != nil {
		s.logger.Log("%s: failed err: %v", op, err)
		return
	}
	s.logger.Log("%s: success", op)
	s.longPoll = longpoll
	s.engine.GET("/queue/subscribe", s.tokenChecker.Handler, func(c *gin.Context) {
		longpoll.SubscriptionHandler(c.Writer, c.Request)
	})
}

func (s *defaultService) CounterpartiesUpdated(uid UserId) {
	type Payload struct{}
	key := fmt.Sprintf("counterparties_%s", uid)
	payload := Payload{}
	s.longPoll.Publish(key, payload)
}

func (s *defaultService) SpendingsUpdated(uid UserId) {
	type Payload struct{}
	key := fmt.Sprintf("spendings_%s", uid)
	payload := Payload{}
	s.longPoll.Publish(key, payload)
}

func (s *defaultService) FriendsUpdated(uid UserId) {
	type Payload struct{}
	key := fmt.Sprintf("friends_%s", uid)
	payload := Payload{}
	s.longPoll.Publish(key, payload)
}
