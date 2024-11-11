package friends

import (
	friendsController "verni/internal/controllers/friends"
	httpserver "verni/internal/http-server"
	"verni/internal/services/logging"
	"verni/internal/services/longpoll"
	"verni/internal/services/pushNotifications"
)

type AcceptFriendRequest struct {
	Sender httpserver.UserId `json:"sender"`
}

type GetFriendsRequest struct {
	Statuses []httpserver.FriendStatus `json:"statuses"`
}

type RejectFriendRequest struct {
	Sender httpserver.UserId `json:"sender"`
}

type RollbackFriendRequest struct {
	Target httpserver.UserId `json:"target"`
}

type SendFriendRequest struct {
	Target httpserver.UserId `json:"target"`
}

type UnfriendRequest struct {
	Target httpserver.UserId `json:"target"`
}

type RequestsHandler interface {
	AcceptRequest(
		subject httpserver.UserId,
		request AcceptFriendRequest,
		success func(httpserver.StatusCode, httpserver.VoidResponse),
		failure func(httpserver.StatusCode, httpserver.Response[httpserver.Error]),
	)
	GetFriends(
		subject httpserver.UserId,
		request GetFriendsRequest,
		success func(httpserver.StatusCode, httpserver.Response[map[httpserver.FriendStatus][]httpserver.UserId]),
		failure func(httpserver.StatusCode, httpserver.Response[httpserver.Error]),
	)
	RejectRequest(
		subject httpserver.UserId,
		request RejectFriendRequest,
		success func(httpserver.StatusCode, httpserver.VoidResponse),
		failure func(httpserver.StatusCode, httpserver.Response[httpserver.Error]),
	)
	RollbackRequest(
		subject httpserver.UserId,
		request RollbackFriendRequest,
		success func(httpserver.StatusCode, httpserver.VoidResponse),
		failure func(httpserver.StatusCode, httpserver.Response[httpserver.Error]),
	)
	SendRequest(
		subject httpserver.UserId,
		request SendFriendRequest,
		success func(httpserver.StatusCode, httpserver.VoidResponse),
		failure func(httpserver.StatusCode, httpserver.Response[httpserver.Error]),
	)
	Unfriend(
		subject httpserver.UserId,
		request UnfriendRequest,
		success func(httpserver.StatusCode, httpserver.VoidResponse),
		failure func(httpserver.StatusCode, httpserver.Response[httpserver.Error]),
	)
}

func DefaultHandler(
	controller friendsController.Controller,
	pushService pushNotifications.Service,
	pollingService longpoll.Service,
	logger logging.Service,
) RequestsHandler {
	return &defaultRequestsHandler{
		controller:     controller,
		pushService:    pushService,
		pollingService: pollingService,
		logger:         logger,
	}
}
