package users

import (
	"github.com/rzmn/Verni.App.Backend/internal/common"
	friendsRepository "github.com/rzmn/Verni.App.Backend/internal/repositories/friends"
)

type UserId string
type AvatarId string
type FriendStatus friendsRepository.FriendStatus

type User struct {
	Id           UserId
	DisplayName  string
	AvatarId     *AvatarId
	FriendStatus FriendStatus
}

type Controller interface {
	Get(ids []UserId, sender UserId) ([]User, *common.CodeBasedError[GetUsersErrorCode])
}
