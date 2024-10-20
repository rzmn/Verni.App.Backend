package friends

import (
	"verni/internal/common"
	friendsRepository "verni/internal/repositories/friends"
	"verni/internal/storage"
)

type UserId storage.UserId
type FriendStatus int
type Repository friendsRepository.Repository

const (
	_ FriendStatus = iota
	FriendStatusFriends
	FriendStatusSubscription
	FriendStatusSubscriber
)

type Controller interface {
	AcceptFriendRequest(sender UserId, target UserId) *common.CodeBasedError[AcceptFriendRequestErrorCode]
	GetFriends(statuses []FriendStatus, userId UserId) (map[FriendStatus][]UserId, *common.CodeBasedError[GetFriendsErrorCode])
	RollbackFriendRequest(sender UserId, target UserId) *common.CodeBasedError[RollbackFriendRequestErrorCode]
	SendFriendRequest(sender UserId, target UserId) *common.CodeBasedError[SendFriendRequestErrorCode]
	Unfriend(sender UserId, target UserId) *common.CodeBasedError[UnfriendErrorCode]
}

func DefaultController(repository Repository) Controller {
	return &defaultController{
		repository: repository,
	}
}
