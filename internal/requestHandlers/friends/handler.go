package friends

import (
	friendsController "verni/internal/controllers/friends"
	httpserver "verni/internal/http-server"
	"verni/internal/http-server/longpoll"
	"verni/internal/http-server/responses"
	"verni/internal/services/logging"
	"verni/internal/services/pushNotifications"
)

type HttpCode int

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
		subject friendsController.UserId,
		request AcceptFriendRequest,
		success func(HttpCode, responses.VoidResponse),
		failure func(HttpCode, responses.Response[responses.Error]),
	)
	GetFriends(
		subject friendsController.UserId,
		request GetFriendsRequest,
		success func(HttpCode, responses.Response[map[httpserver.FriendStatus][]httpserver.UserId]),
		failure func(HttpCode, responses.Response[responses.Error]),
	)
	RejectRequest(
		subject friendsController.UserId,
		request RejectFriendRequest,
		success func(HttpCode, responses.VoidResponse),
		failure func(HttpCode, responses.Response[responses.Error]),
	)
	RollbackRequest(
		subject friendsController.UserId,
		request RollbackFriendRequest,
		success func(HttpCode, responses.VoidResponse),
		failure func(HttpCode, responses.Response[responses.Error]),
	)
	SendRequest(
		subject friendsController.UserId,
		request SendFriendRequest,
		success func(HttpCode, responses.VoidResponse),
		failure func(HttpCode, responses.Response[responses.Error]),
	)
	Unfriend(
		subject friendsController.UserId,
		request UnfriendRequest,
		success func(HttpCode, responses.VoidResponse),
		failure func(HttpCode, responses.Response[responses.Error]),
	)
}

func DefaultHandler(
	controller friendsController.Controller,
	pushService pushNotifications.Service,
	pollingService longpoll.Service,
	logger logging.Service,
) RequestsHandler {
	return &defaultRequestsHandler{}
}
