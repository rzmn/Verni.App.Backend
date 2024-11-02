package users

import (
	"verni/internal/common"
	friendsRepository "verni/internal/repositories/friends"
	usersRepository "verni/internal/repositories/users"
)

type UserId string
type AvatarId string
type UsersRepository usersRepository.Repository
type FriendsRepository friendsRepository.Repository
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

func DefaultController(users UsersRepository, friends FriendsRepository) Controller {
	return &defaultController{
		users:   users,
		friends: friends,
	}
}
