package friends

import (
	"verni/internal/repositories"
)

type UserId string

type FriendStatus int

const (
	FriendStatusNo = iota
	FriendStatusSubscriber
	FriendStatusSubscription
	FriendStatusFriend
	FriendStatusMe
)

type Repository interface {
	GetFriends(userId UserId) ([]UserId, error)
	GetSubscribers(userId UserId) ([]UserId, error)
	GetSubscriptions(userId UserId) ([]UserId, error)
	GetStatuses(sender UserId, ids []UserId) (map[UserId]FriendStatus, error)
	HasFriendRequest(sender UserId, target UserId) (bool, error)

	StoreFriendRequest(sender UserId, target UserId) repositories.MutationWorkItem
	RemoveFriendRequest(sender UserId, target UserId) repositories.MutationWorkItem
}
