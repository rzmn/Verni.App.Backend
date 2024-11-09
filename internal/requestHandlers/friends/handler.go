package friends

import (
	friendsController "verni/internal/controllers/friends"
	httpserver "verni/internal/http-server"
	"verni/internal/http-server/longpoll"
	"verni/internal/http-server/responses"
	"verni/internal/services/logging"
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
		success func(httpserver.StatusCode, responses.VoidResponse),
		failure func(httpserver.StatusCode, responses.Response[responses.Error]),
	)
	GetFriends(
		subject httpserver.UserId,
		request GetFriendsRequest,
		success func(httpserver.StatusCode, responses.Response[map[httpserver.FriendStatus][]httpserver.UserId]),
		failure func(httpserver.StatusCode, responses.Response[responses.Error]),
	)
	RejectRequest(
		subject httpserver.UserId,
		request RejectFriendRequest,
		success func(httpserver.StatusCode, responses.VoidResponse),
		failure func(httpserver.StatusCode, responses.Response[responses.Error]),
	)
	RollbackRequest(
		subject httpserver.UserId,
		request RollbackFriendRequest,
		success func(httpserver.StatusCode, responses.VoidResponse),
		failure func(httpserver.StatusCode, responses.Response[responses.Error]),
	)
	SendRequest(
		subject httpserver.UserId,
		request SendFriendRequest,
		success func(httpserver.StatusCode, responses.VoidResponse),
		failure func(httpserver.StatusCode, responses.Response[responses.Error]),
	)
	Unfriend(
		subject httpserver.UserId,
		request UnfriendRequest,
		success func(httpserver.StatusCode, responses.VoidResponse),
		failure func(httpserver.StatusCode, responses.Response[responses.Error]),
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
