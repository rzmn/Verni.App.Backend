package friends

import (
	friendsController "verni/internal/controllers/friends"
	"verni/internal/schema"
	"verni/internal/services/logging"
	"verni/internal/services/longpoll"
	"verni/internal/services/pushNotifications"
)

type AcceptFriendRequest struct {
	Sender schema.UserId `json:"sender"`
}

type GetFriendsRequest struct {
	Statuses []schema.FriendStatus `json:"statuses"`
}

type RejectFriendRequest struct {
	Sender schema.UserId `json:"sender"`
}

type RollbackFriendRequest struct {
	Target schema.UserId `json:"target"`
}

type SendFriendRequest struct {
	Target schema.UserId `json:"target"`
}

type UnfriendRequest struct {
	Target schema.UserId `json:"target"`
}

type RequestsHandler interface {
	AcceptRequest(
		subject schema.UserId,
		request AcceptFriendRequest,
		success func(schema.StatusCode, schema.VoidResponse),
		failure func(schema.StatusCode, schema.Response[schema.Error]),
	)
	GetFriends(
		subject schema.UserId,
		request GetFriendsRequest,
		success func(schema.StatusCode, schema.Response[map[schema.FriendStatus][]schema.UserId]),
		failure func(schema.StatusCode, schema.Response[schema.Error]),
	)
	RejectRequest(
		subject schema.UserId,
		request RejectFriendRequest,
		success func(schema.StatusCode, schema.VoidResponse),
		failure func(schema.StatusCode, schema.Response[schema.Error]),
	)
	RollbackRequest(
		subject schema.UserId,
		request RollbackFriendRequest,
		success func(schema.StatusCode, schema.VoidResponse),
		failure func(schema.StatusCode, schema.Response[schema.Error]),
	)
	SendRequest(
		subject schema.UserId,
		request SendFriendRequest,
		success func(schema.StatusCode, schema.VoidResponse),
		failure func(schema.StatusCode, schema.Response[schema.Error]),
	)
	Unfriend(
		subject schema.UserId,
		request UnfriendRequest,
		success func(schema.StatusCode, schema.VoidResponse),
		failure func(schema.StatusCode, schema.Response[schema.Error]),
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
