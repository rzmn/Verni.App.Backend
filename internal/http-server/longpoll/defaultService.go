package longpoll

import (
	"fmt"
	"log"
	"verni/internal/auth/jwt"
	"verni/internal/http-server/middleware"
	"verni/internal/storage"

	"github.com/jcuga/golongpoll"

	"github.com/gin-gonic/gin"
)

type defaultService struct {
	engine     *gin.Engine
	db         storage.Storage
	jwtService jwt.Service
	longPoll   *golongpoll.LongpollManager
}

func (s *defaultService) RegisterRoutes() {
	const op = "longpoll.defaultService.RegisterRoutes"
	log.Printf("%s: start", op)
	longpoll, err := golongpoll.StartLongpoll(golongpoll.Options{})
	if err != nil {
		log.Printf("%s: failed err: %v", op, err)
		return
	}
	log.Printf("%s: success", op)
	s.longPoll = longpoll
	ensureLoggedIn := middleware.EnsureLoggedIn(s.db, s.jwtService)
	s.engine.GET("/queue/subscribe", ensureLoggedIn, func(c *gin.Context) {
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
