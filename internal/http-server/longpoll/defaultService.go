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

func (c *defaultService) RegisterRoutes() {
	const op = "longpoll.defaultService.RegisterRoutes"
	c.logger.LogInfo("%s: start", op)
	longpoll, err := golongpoll.StartLongpoll(golongpoll.Options{})
	if err != nil {
		c.logger.LogInfo("%s: failed err: %v", op, err)
		return
	}
	c.logger.LogInfo("%s: success", op)
	c.longPoll = longpoll
	c.engine.GET("/queue/subscribe", c.tokenChecker.Handler, func(c *gin.Context) {
		longpoll.SubscriptionHandler(c.Writer, c.Request)
	})
}

func (c *defaultService) CounterpartiesUpdated(uid UserId) {
	type Payload struct{}
	key := fmt.Sprintf("counterparties_%s", uid)
	payload := Payload{}
	c.longPoll.Publish(key, payload)
}

func (c *defaultService) SpendingsUpdated(uid UserId) {
	type Payload struct{}
	key := fmt.Sprintf("spendings_%s", uid)
	payload := Payload{}
	c.longPoll.Publish(key, payload)
}

func (c *defaultService) FriendsUpdated(uid UserId) {
	type Payload struct{}
	key := fmt.Sprintf("friends_%s", uid)
	payload := Payload{}
	c.longPoll.Publish(key, payload)
}
