package friends

import (
	"github.com/rzmn/Verni.App.Backend/internal/common"
)

type UserId string
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
	RollbackFriendRequest(sender UserId, target UserId) *common.CodeBasedError[RollbackFriendRequestErrorCode]
	SendFriendRequest(sender UserId, target UserId) *common.CodeBasedError[SendFriendRequestErrorCode]
	Unfriend(sender UserId, target UserId) *common.CodeBasedError[UnfriendErrorCode]
}
