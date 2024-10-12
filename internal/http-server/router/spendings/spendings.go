package spendings

import (
	"accounty/internal/apns"
	"accounty/internal/auth/jwt"
	"accounty/internal/http-server/handlers/spendings/createDeal"
	"accounty/internal/http-server/handlers/spendings/deleteDeal"
	"accounty/internal/http-server/handlers/spendings/getCounterparties"
	"accounty/internal/http-server/handlers/spendings/getDeal"
	"accounty/internal/http-server/handlers/spendings/getDeals"
	"accounty/internal/http-server/middleware"
	"accounty/internal/storage"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jcuga/golongpoll"
)

type createDealRequestHandler struct {
	storage    storage.Storage
	pushSender apns.PushNotificationSender
	longPoll   *golongpoll.LongpollManager
}

func (h *createDealRequestHandler) Validate(c *gin.Context, request createDeal.Request) *createDeal.Error {
	const op = "router.friends.createDealRequestHandler.Validate"
	log.Printf("%s: start with request %v", op, request)
	subject := storage.UserId(c.Request.Header.Get(middleware.LoggedInSubjectKey))
	for i := 0; i < len(request.Deal.Spendings); i++ {
		if request.Deal.Spendings[i].UserId == subject {
			continue
		}
		exists, err := h.storage.IsUserExists(request.Deal.Spendings[i].UserId)
		if err != nil {
			outError := createDeal.ErrInternal()
			return &outError
		}
		if !exists {
			outError := createDeal.ErrNoSuchUser()
			return &outError
		}
		has, err := h.storage.HasFriendship(subject, request.Deal.Spendings[i].UserId)
		if err != nil {
			outError := createDeal.ErrInternal()
			return &outError
		}
		if !has {
			outError := createDeal.ErrNotAFriend()
			return &outError
		}
	}
	return nil
}

func (h *createDealRequestHandler) Handle(c *gin.Context, request createDeal.Request) *createDeal.Error {
	const op = "router.friends.createDealRequestHandler.Handle"
	log.Printf("%s: start with request %v", op, request)
	dealId, err := h.storage.InsertDeal(request.Deal)
	if err != nil {
		outError := createDeal.ErrInternal()
		return &outError
	}
	subject := storage.UserId(c.Request.Header.Get(middleware.LoggedInSubjectKey))
	h.longPoll.Publish(LongPollCounterpartiesUpdateKey(), storage.LongPollUpdatePayload{})
	for i := 0; i < len(request.Deal.Spendings); i++ {
		spending := request.Deal.Spendings[i]
		h.longPoll.Publish(LongPollSpendingsHistoryUpdateKey(spending.UserId), storage.LongPollUpdatePayload{})
		if spending.UserId == subject {
			continue
		}
		receiverToken, err := h.storage.GetPushToken(spending.UserId)
		if err != nil {
			log.Printf("%s: cannot get receiver push token info %v", op, err)
		} else if receiverToken == nil {
			log.Printf("%s: receiver push token is nil", op)
		} else {
			h.pushSender.NewExpenseReceived(*receiverToken, storage.IdentifiableDeal{
				Id:   dealId,
				Deal: request.Deal,
			}, subject, spending.UserId)
		}
	}
	return nil
}

type deleteDealRequestHandler struct {
	storage  storage.Storage
	longPoll *golongpoll.LongpollManager
}

func (h *deleteDealRequestHandler) Validate(c *gin.Context, request deleteDeal.Request) *deleteDeal.Error {
	const op = "router.friends.deleteDealRequestHandler.Validate"
	log.Printf("%s: start with request %v", op, request)
	subject := storage.UserId(c.Request.Header.Get(middleware.LoggedInSubjectKey))
	dealFromDb, err := h.storage.GetDeal(request.DealId)
	if err != nil {
		outError := deleteDeal.ErrInternal()
		return &outError
	}
	if dealFromDb == nil {
		outError := deleteDeal.ErrDealNotFound()
		return &outError
	}
	counterparties, err := h.storage.GetCounterpartiesForDeal(request.DealId)
	if err != nil {
		outError := deleteDeal.ErrInternal()
		return &outError
	}
	var isYourDeal bool
	for i := 0; i < len(counterparties); i++ {
		if counterparties[i] == subject {
			isYourDeal = true
		} else {
			isFriend, err := h.storage.HasFriendship(subject, counterparties[i])
			if err != nil {
				outError := deleteDeal.ErrInternal()
				return &outError
			}
			if !isFriend {
				outError := deleteDeal.ErrNotAFriend()
				return &outError
			}
		}
	}
	if !isYourDeal {
		outError := deleteDeal.ErrIsNotYourDeal()
		return &outError
	}
	return nil
}

func (h *deleteDealRequestHandler) Handle(c *gin.Context, request deleteDeal.Request) *deleteDeal.Error {
	const op = "router.friends.deleteDealRequestHandler.Handle"
	log.Printf("%s: start with request %v", op, request)
	deal, getDealErr := h.storage.GetDeal(request.DealId)
	if err := h.storage.RemoveDeal(request.DealId); err != nil {
		outError := deleteDeal.ErrInternal()
		return &outError
	}
	h.longPoll.Publish(LongPollCounterpartiesUpdateKey(), storage.LongPollUpdatePayload{})
	if getDealErr == nil && deal != nil {
		for i := 0; i < len(deal.Spendings); i++ {
			h.longPoll.Publish(LongPollSpendingsHistoryUpdateKey(deal.Spendings[i].UserId), storage.LongPollUpdatePayload{})
		}
	}
	return nil
}

type getCounterpartiesRequestHandler struct {
	storage storage.Storage
}

func (h *getCounterpartiesRequestHandler) Handle(c *gin.Context, request getCounterparties.Request) ([]storage.SpendingsPreview, *getCounterparties.Error) {
	const op = "router.friends.getCounterpartiesRequestHandler.Handle"
	log.Printf("%s: start with request %v", op, request)
	subject := storage.UserId(c.Request.Header.Get(middleware.LoggedInSubjectKey))
	preview, err := h.storage.GetCounterparties(subject)
	if err != nil {
		outError := getCounterparties.ErrInternal()
		return []storage.SpendingsPreview{}, &outError
	}
	return preview, nil
}

type getDealsRequestHandler struct {
	storage storage.Storage
}

func (h *getDealsRequestHandler) Validate(c *gin.Context, request getDeals.Request) *getDeals.Error {
	const op = "router.spendings.getDealsRequestHandler.Validate"
	log.Printf("%s: start with request %v", op, request)
	exists, err := h.storage.IsUserExists(request.Counterparty)
	if err != nil {
		outError := getDeals.ErrInternal()
		return &outError
	}
	if !exists {
		outError := getDeals.ErrNoSuchUser()
		return &outError
	}
	return nil
}

func (h *getDealsRequestHandler) Handle(c *gin.Context, request getDeals.Request) ([]storage.IdentifiableDeal, *getDeals.Error) {
	const op = "router.spendings.getDealsRequestHandler.Handle"
	log.Printf("%s: start with request %v", op, request)
	subject := storage.UserId(c.Request.Header.Get(middleware.LoggedInSubjectKey))
	deals, err := h.storage.GetDeals(request.Counterparty, subject)
	if err != nil {
		outError := getDeals.ErrInternal()
		return []storage.IdentifiableDeal{}, &outError
	}
	return deals, nil
}

type getDealRequestHandler struct {
	storage storage.Storage
}

func (h *getDealRequestHandler) Handle(c *gin.Context, request getDeal.Request) (storage.Deal, *getDeal.Error) {
	const op = "router.spendings.getDealRequestHandler.Handle"
	log.Printf("%s: start with request %v", op, request)
	deal, err := h.storage.GetDeal(request.Id)
	if err != nil {
		outError := getDeal.ErrInternal()
		return storage.Deal{}, &outError
	}
	if deal == nil {
		outError := getDeal.ErrDealNotFound()
		return storage.Deal{}, &outError
	}
	return deal.Deal, nil
}

func LongPollCounterpartiesUpdateKey() string {
	return "counterparties"
}

func LongPollSpendingsHistoryUpdateKey(uid storage.UserId) string {
	return fmt.Sprintf("spendings_%s", uid)
}

func RegisterRoutes(e *gin.Engine, storage storage.Storage, pushSender apns.PushNotificationSender, jwtService jwt.Service) {
	longpoll, err := golongpoll.StartLongpoll(golongpoll.Options{})
	if err != nil {
		panic(err)
	}
	group := e.Group("/spendings", middleware.EnsureLoggedIn(storage, jwtService))
	group.POST("/createDeal", createDeal.New(&createDealRequestHandler{storage: storage, pushSender: pushSender, longPoll: longpoll}))
	group.POST("/deleteDeal", deleteDeal.New(&deleteDealRequestHandler{storage: storage, longPoll: longpoll}))
	group.GET("/getCounterparties", getCounterparties.New(&getCounterpartiesRequestHandler{storage: storage}))
	group.GET("/getDeals", getDeals.New(&getDealsRequestHandler{storage: storage}))
	group.GET("/getDeal", getDeal.New(&getDealRequestHandler{storage: storage}))

	group.GET("/subscribe", wrapWithContext(longpoll.SubscriptionHandler))
}

func wrapWithContext(lpHandler func(http.ResponseWriter, *http.Request)) func(*gin.Context) {
	return func(c *gin.Context) {
		lpHandler(c.Writer, c.Request)
	}
}
