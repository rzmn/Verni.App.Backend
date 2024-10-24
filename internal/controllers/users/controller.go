package users

import (
	"verni/internal/common"
	friendsRepository "verni/internal/repositories/friends"
	usersRepository "verni/internal/repositories/users"
)

type UserId string
type AvatarId string
type Repository usersRepository.Repository
type FriendStatus friendsRepository.FriendStatus

type User struct {
	Id           UserId
	DisplayName  string
	AvatarId     *AvatarId
	FriendStatus FriendStatus
}

type Controller interface {
	Get(ids []UserId, sender UserId) ([]User, *common.CodeBasedError[GetUsersErrorCode])
	Search(query string, sender UserId) ([]User, *common.CodeBasedError[SearchUsersErrorCode])
}

func DefaultController(repository Repository) Controller {
	return &defaultController{
		repository: repository,
	}
}
