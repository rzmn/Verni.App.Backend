package longpoll

import (
	"fmt"
	"verni/internal/services/logging"

	"github.com/jcuga/golongpoll"
)

type ginService struct {
	longPoll *golongpoll.LongpollManager
	logger   logging.Service
}

func (c *ginService) CounterpartiesUpdated(uid UserId) {
	op := "longpoll.CounterpartiesUpdated"
	c.logger.LogInfo("%s: start[uid=%s]", op, uid)
	type Payload struct{}
	key := fmt.Sprintf("counterparties_%s", uid)
	payload := Payload{}
	c.longPoll.Publish(key, payload)
	c.logger.LogInfo("%s: success[uid=%s]", op, uid)
}

func (c *ginService) ExpensesUpdated(uid UserId, counterparty UserId) {
	op := "longpoll.ExpensesUpdated"
	c.logger.LogInfo("%s: start[uid=%s, cid=%s]", op, uid, counterparty)
	type Payload struct{}
	key := fmt.Sprintf("spendings_%s_%s", uid, counterparty)
	payload := Payload{}
	c.longPoll.Publish(key, payload)
	c.logger.LogInfo("%s: success[uid=%s, cid=%s]", op, uid, counterparty)
}

func (c *ginService) FriendsUpdated(uid UserId) {
	op := "longpoll.FriendsUpdated"
	c.logger.LogInfo("%s: start[uid=%s]", op, uid)
	type Payload struct{}
	key := fmt.Sprintf("friends_%s", uid)
	payload := Payload{}
	c.longPoll.Publish(key, payload)
	c.logger.LogInfo("%s: success[uid=%s]", op, uid)
}
