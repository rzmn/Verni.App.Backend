package ginLongpollRealtimeEvents

import (
	"fmt"
	"verni/internal/services/logging"
	"verni/internal/services/realtimeEvents"

	"github.com/gin-gonic/gin"
	"github.com/jcuga/golongpoll"
)

func New(
	e *gin.Engine,
	logger logging.Service,
	accessTokenMiddleware gin.HandlerFunc,
) realtimeEvents.Service {
	op := "realtimeEvents.GinService"
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

type ginService struct {
	longPoll *golongpoll.LongpollManager
	logger   logging.Service
}

func (c *ginService) CounterpartiesUpdated(uid realtimeEvents.UserId) {
	op := "longpoll.CounterpartiesUpdated"
	c.logger.LogInfo("%s: start[uid=%s]", op, uid)
	type Payload struct{}
	key := fmt.Sprintf("counterparties_%s", uid)
	payload := Payload{}
	c.longPoll.Publish(key, payload)
	c.logger.LogInfo("%s: success[uid=%s]", op, uid)
}

func (c *ginService) ExpensesUpdated(uid realtimeEvents.UserId, counterparty realtimeEvents.UserId) {
	op := "longpoll.ExpensesUpdated"
	c.logger.LogInfo("%s: start[uid=%s, cid=%s]", op, uid, counterparty)
	type Payload struct{}
	key := fmt.Sprintf("spendings_%s_%s", uid, counterparty)
	payload := Payload{}
	c.longPoll.Publish(key, payload)
	c.logger.LogInfo("%s: success[uid=%s, cid=%s]", op, uid, counterparty)
}

func (c *ginService) FriendsUpdated(uid realtimeEvents.UserId) {
	op := "longpoll.FriendsUpdated"
	c.logger.LogInfo("%s: start[uid=%s]", op, uid)
	type Payload struct{}
	key := fmt.Sprintf("friends_%s", uid)
	payload := Payload{}
	c.longPoll.Publish(key, payload)
	c.logger.LogInfo("%s: success[uid=%s]", op, uid)
}
