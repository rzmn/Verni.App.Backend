package friends

import (
	"log"
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/jcuga/golongpoll"

	"accounty/internal/apns"
	"accounty/internal/auth/jwt"
	"accounty/internal/storage"

	"accounty/internal/http-server/handlers/friends/acceptRequest"
	"accounty/internal/http-server/handlers/friends/get"
	"accounty/internal/http-server/handlers/friends/rejectRequest"
	"accounty/internal/http-server/handlers/friends/rollbackRequest"
	"accounty/internal/http-server/handlers/friends/sendRequest"
	"accounty/internal/http-server/handlers/friends/unfriend"
	"accounty/internal/http-server/middleware"
)

type sendRequestRequestHandler struct {
	storage    storage.Storage
	pushSender apns.PushNotificationSender
	longPoll   *golongpoll.LongpollManager
}

func (h *sendRequestRequestHandler) Validate(c *gin.Context, request sendRequest.Request) *sendRequest.Error {
	const op = "router.sendRequestRequestHandler.Validate"
	log.Printf("%s: start with request %v", op, request)
	subject := storage.UserId(c.Request.Header.Get(middleware.LoggedInSubjectKey))
	hasTarget, err := h.storage.IsUserExists(request.Target)
	if err != nil {
		outError := sendRequest.ErrInternal()
		return &outError
	} else if !hasTarget {
		outError := sendRequest.ErrNoSuchUser()
		return &outError
	}
	hasRequest, err := h.storage.HasFriendRequest(subject, request.Target)
	if err != nil {
		outError := sendRequest.ErrInternal()
		return &outError
	} else if hasRequest {
		outError := sendRequest.ErrAlreadySend()
		return &outError
	}
	hasIncomingRequest, err := h.storage.HasFriendRequest(request.Target, subject)
	if err != nil {
		outError := sendRequest.ErrInternal()
		return &outError
	} else if hasIncomingRequest {
		outError := sendRequest.ErrHaveIncomingRequest()
		return &outError
	}
	isFriends, err := h.storage.HasFriendship(request.Target, subject)
	if err != nil {
		outError := sendRequest.ErrInternal()
		return &outError
	} else if isFriends {
		outError := sendRequest.ErrAlreadyFriends()
		return &outError
	}
	return nil
}

func (h *sendRequestRequestHandler) Handle(c *gin.Context, request sendRequest.Request) *sendRequest.Error {
	const op = "router.friends.sendRequestRequestHandler.Handle"
	log.Printf("%s: start with request %v", op, request)
	subject := storage.UserId(c.Request.Header.Get(middleware.LoggedInSubjectKey))
	if err := h.storage.StoreFriendRequest(subject, request.Target); err != nil {
		outError := sendRequest.ErrInternal()
		return &outError
	}
	profile, err := h.storage.GetAccountInfo(subject)
	if err != nil {
		log.Printf("%s: cannot get account info for push notification %v", op, err)
	} else {
		receiverToken, err := h.storage.GetPushToken(request.Target)
		if err != nil {
			log.Printf("%s: cannot get receiver push token info %v", op, err)
		} else if receiverToken == nil {
			log.Printf("%s: receiver push token is nil", op)
		} else {
			h.pushSender.GotFriendRequest(*receiverToken, profile.User.Id)
		}
	}
	h.longPoll.Publish(LongPollFriendsUpdateKey(), storage.LongPollUpdatePayload{})
	return nil
}

type acceptRequestRequestHandler struct {
	storage    storage.Storage
	pushSender apns.PushNotificationSender
	longPoll   *golongpoll.LongpollManager
}

func (h *acceptRequestRequestHandler) Validate(c *gin.Context, request acceptRequest.Request) *acceptRequest.Error {
	const op = "router.friends.acceptRequestRequestHandler.Validate"
	log.Printf("%s: start with request %v", op, request)
	subject := storage.UserId(c.Request.Header.Get(middleware.LoggedInSubjectKey))
	hasRequest, err := h.storage.HasFriendRequest(request.Sender, subject)
	if err != nil {
		outError := acceptRequest.ErrInternal()
		return &outError
	} else if !hasRequest {
		outError := acceptRequest.ErrNoSuchRequest()
		return &outError
	}
	return nil
}

func (h *acceptRequestRequestHandler) Handle(c *gin.Context, request acceptRequest.Request) *acceptRequest.Error {
	const op = "router.friends.acceptRequestRequestHandler.Handle"
	log.Printf("%s: start with request %v", op, request)
	subject := storage.UserId(c.Request.Header.Get(middleware.LoggedInSubjectKey))
	if err := h.storage.RemoveFriendRequest(request.Sender, subject); err != nil {
		outError := acceptRequest.ErrInternal()
		return &outError
	}
	if err := h.storage.StoreFriendship(request.Sender, subject); err != nil {
		outError := acceptRequest.ErrInternal()
		return &outError
	}
	profile, err := h.storage.GetAccountInfo(subject)
	if err != nil {
		log.Printf("%s: cannot get account info for push notification %v", op, err)
	} else {
		senderToken, err := h.storage.GetPushToken(request.Sender)
		if err != nil {
			log.Printf("%s: cannot get sender push token info %v", op, err)
		} else if senderToken == nil {
			log.Printf("%s: sender push token is nil", op)
		} else {
			h.pushSender.FriendRequestHasBeenAccepted(*senderToken, profile.User.Id)
		}
	}
	h.longPoll.Publish(LongPollFriendsUpdateKey(), storage.LongPollUpdatePayload{})
	return nil
}

type getRequestHandler struct {
	storage storage.Storage
}

func (h *getRequestHandler) Handle(c *gin.Context, request get.Request) (map[get.Status][]storage.UserId, *get.Error) {
	const op = "router.friends.getRequestHandler.Handle"
	log.Printf("%s: start with request %v", op, request)
	subject := storage.UserId(c.Request.Header.Get(middleware.LoggedInSubjectKey))
	friends := map[get.Status][]storage.UserId{}
	if slices.Contains(request.Statuses, get.FriendStatusFriends) {
		logins, err := h.storage.GetFriends(subject)
		if err != nil {
			outError := get.ErrInternal()
			return map[get.Status][]storage.UserId{}, &outError
		}
		for i := range logins {
			friends[get.FriendStatusFriends] = append(friends[get.FriendStatusFriends], storage.UserId(logins[i]))
		}
	}
	if slices.Contains(request.Statuses, get.FriendStatusSubscriber) {
		logins, err := h.storage.GetIncomingRequests(subject)
		if err != nil {
			outError := get.ErrInternal()
			return map[get.Status][]storage.UserId{}, &outError
		}
		for i := range logins {
			friends[get.FriendStatusSubscriber] = append(friends[get.FriendStatusSubscriber], storage.UserId(logins[i]))
		}
	}
	if slices.Contains(request.Statuses, get.FriendStatusSubscription) {
		logins, err := h.storage.GetPendingRequests(subject)
		if err != nil {
			outError := get.ErrInternal()
			return map[get.Status][]storage.UserId{}, &outError
		}
		for i := range logins {
			friends[get.FriendStatusSubscription] = append(friends[get.FriendStatusSubscription], storage.UserId(logins[i]))
		}
	}
	return friends, nil
}

type rejectRequestRequestHandler struct {
	storage  storage.Storage
	longPoll *golongpoll.LongpollManager
}

func (h *rejectRequestRequestHandler) Validate(c *gin.Context, request rejectRequest.Request) *rejectRequest.Error {
	const op = "router.friends.rejectRequestRequestHandler.Validate"
	log.Printf("%s: start with request %v", op, request)
	subject := storage.UserId(c.Request.Header.Get(middleware.LoggedInSubjectKey))
	hasRequest, err := h.storage.HasFriendRequest(request.Sender, subject)
	if err != nil {
		outError := rejectRequest.ErrInternal()
		return &outError
	} else if !hasRequest {
		outError := rejectRequest.ErrNoSuchRequest()
		return &outError
	}
	return nil
}

func (h *rejectRequestRequestHandler) Handle(c *gin.Context, request rejectRequest.Request) *rejectRequest.Error {
	const op = "router.friends.acceptRequestRequestHandler.Handle"
	log.Printf("%s: start with request %v", op, request)
	subject := storage.UserId(c.Request.Header.Get(middleware.LoggedInSubjectKey))
	if err := h.storage.RemoveFriendRequest(request.Sender, subject); err != nil {
		outError := rejectRequest.ErrInternal()
		return &outError
	}
	h.longPoll.Publish(LongPollFriendsUpdateKey(), storage.LongPollUpdatePayload{})
	return nil
}

type rollbackRequestRequestHandler struct {
	storage  storage.Storage
	longPoll *golongpoll.LongpollManager
}

func (h *rollbackRequestRequestHandler) Validate(c *gin.Context, request rollbackRequest.Request) *rollbackRequest.Error {
	const op = "router.friends.rollbackRequestRequestHandler.Validate"
	log.Printf("%s: start with request %v", op, request)
	subject := storage.UserId(c.Request.Header.Get(middleware.LoggedInSubjectKey))
	hasRequest, err := h.storage.HasFriendRequest(subject, request.Target)
	if err != nil {
		outError := rollbackRequest.ErrInternal()
		return &outError
	} else if !hasRequest {
		outError := rollbackRequest.ErrNoSuchRequest()
		return &outError
	}
	return nil
}

func (h *rollbackRequestRequestHandler) Handle(c *gin.Context, request rollbackRequest.Request) *rollbackRequest.Error {
	const op = "router.friends.rollbackRequestRequestHandler.Handle"
	log.Printf("%s: start with request %v", op, request)
	subject := storage.UserId(c.Request.Header.Get(middleware.LoggedInSubjectKey))
	if err := h.storage.RemoveFriendRequest(subject, request.Target); err != nil {
		outError := rollbackRequest.ErrInternal()
		return &outError
	}
	h.longPoll.Publish(LongPollFriendsUpdateKey(), storage.LongPollUpdatePayload{})
	return nil
}

type unfriendRequestHandler struct {
	storage  storage.Storage
	longPoll *golongpoll.LongpollManager
}

func (h *unfriendRequestHandler) Validate(c *gin.Context, request unfriend.Request) *unfriend.Error {
	const op = "router.friends.unfriendRequestHandler.Validate"
	log.Printf("%s: start with request %v", op, request)
	subject := storage.UserId(c.Request.Header.Get(middleware.LoggedInSubjectKey))
	hasTarget, err := h.storage.IsUserExists(request.Target)
	if err != nil {
		outError := unfriend.ErrInternal()
		return &outError
	} else if !hasTarget {
		outError := unfriend.ErrNoSuchUser()
		return &outError
	}
	isFriends, err := h.storage.HasFriendship(request.Target, subject)
	if err != nil {
		outError := unfriend.ErrInternal()
		return &outError
	} else if !isFriends {
		outError := unfriend.ErrNotAFriend()
		return &outError
	}
	return nil
}

func (h *unfriendRequestHandler) Handle(c *gin.Context, request unfriend.Request) *unfriend.Error {
	const op = "router.friends.unfriendRequestHandler.Handle"
	log.Printf("%s: start with request %v", op, request)
	subject := storage.UserId(c.Request.Header.Get(middleware.LoggedInSubjectKey))
	if err := h.storage.RemoveFriendship(request.Target, subject); err != nil {
		outError := unfriend.ErrInternal()
		return &outError
	}
	if err := h.storage.StoreFriendRequest(request.Target, subject); err != nil {
		outError := unfriend.ErrInternal()
		return &outError
	}
	h.longPoll.Publish(LongPollFriendsUpdateKey(), storage.LongPollUpdatePayload{})
	return nil
}

func LongPollFriendsUpdateKey() string {
	return "friends"
}

func RegisterRoutes(e *gin.Engine, storage storage.Storage, pushSender apns.PushNotificationSender, jwtService jwt.Service) {
	longpoll, err := golongpoll.StartLongpoll(golongpoll.Options{})
	if err != nil {
		panic(err)
	}
	group := e.Group("/friends", middleware.EnsureLoggedIn(storage, jwtService))
	group.POST("/acceptRequest", acceptRequest.New(&acceptRequestRequestHandler{storage: storage, pushSender: pushSender, longPoll: longpoll}))
	group.GET("/get", get.New(&getRequestHandler{storage: storage}))
	group.POST("/rejectRequest", rejectRequest.New(&rejectRequestRequestHandler{storage: storage, longPoll: longpoll}))
	group.POST("/rollbackRequest", rollbackRequest.New(&rollbackRequestRequestHandler{storage: storage, longPoll: longpoll}))
	group.POST("/sendRequest", sendRequest.New(&sendRequestRequestHandler{storage: storage, pushSender: pushSender, longPoll: longpoll}))
	group.POST("/unfriend", unfriend.New(&unfriendRequestHandler{storage: storage, longPoll: longpoll}))

	group.GET("/subscribe", wrapWithContext(longpoll.SubscriptionHandler))
}

func wrapWithContext(lpHandler func(http.ResponseWriter, *http.Request)) func(*gin.Context) {
	return func(c *gin.Context) {
		lpHandler(c.Writer, c.Request)
	}
}
