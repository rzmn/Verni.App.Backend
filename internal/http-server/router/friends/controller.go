package friends

import (
	"accounty/internal/common"
	"accounty/internal/storage"
)

type UserId storage.UserId
type FriendStatus int

const (
	_ FriendStatus = iota
	FriendStatusFriends
	FriendStatusSubscription
	FriendStatusSubscriber
)

type Controller interface {
	AcceptFriendRequest(sender UserId, target UserId) *common.CodeBasedError[AcceptFriendRequestErrorCode]
	GetFriends(statuses []FriendStatus, userId UserId) (map[FriendStatus][]UserId, *common.CodeBasedError[GetFriendsErrorCode])
	RejectFriendRequest(sender UserId, target UserId) *common.CodeBasedError[RejectFriendRequestErrorCode]
	RollbackFriendRequest(sender UserId, target UserId) *common.CodeBasedError[RollbackFriendRequestErrorCode]
	SendFriendRequest(sender UserId, target UserId) *common.CodeBasedError[SendFriendRequestErrorCode]
	Unfriend(sender UserId, target UserId) *common.CodeBasedError[UnfriendErrorCode]
}

func DefaultController(storage storage.Storage) Controller {
	return &defaultController{
		storage: storage,
	}
}
