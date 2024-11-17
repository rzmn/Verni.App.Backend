package friends

import (
	"verni/internal/schema"
)

type RequestsHandler interface {
	AcceptRequest(
		subject schema.UserId,
		request schema.AcceptFriendRequest,
		success func(schema.StatusCode, schema.VoidResponse),
		failure func(schema.StatusCode, schema.Response[schema.Error]),
	)
	GetFriends(
		subject schema.UserId,
		request schema.GetFriendsRequest,
		success func(schema.StatusCode, schema.Response[map[schema.FriendStatus][]schema.UserId]),
		failure func(schema.StatusCode, schema.Response[schema.Error]),
	)
	RejectRequest(
		subject schema.UserId,
		request schema.RejectFriendRequest,
		success func(schema.StatusCode, schema.VoidResponse),
		failure func(schema.StatusCode, schema.Response[schema.Error]),
	)
	RollbackRequest(
		subject schema.UserId,
		request schema.RollbackFriendRequest,
		success func(schema.StatusCode, schema.VoidResponse),
		failure func(schema.StatusCode, schema.Response[schema.Error]),
	)
	SendRequest(
		subject schema.UserId,
		request schema.SendFriendRequest,
		success func(schema.StatusCode, schema.VoidResponse),
		failure func(schema.StatusCode, schema.Response[schema.Error]),
	)
	Unfriend(
		subject schema.UserId,
		request schema.UnfriendRequest,
		success func(schema.StatusCode, schema.VoidResponse),
		failure func(schema.StatusCode, schema.Response[schema.Error]),
	)
}
