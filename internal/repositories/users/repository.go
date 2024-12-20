package users

import (
	"github.com/rzmn/governi/internal/repositories"
)

type UserId string
type AvatarId string
type User struct {
	Id          UserId
	DisplayName string
	AvatarId    *AvatarId
}

type Repository interface {
	StoreUser(user User) repositories.MutationWorkItem
	GetUsers(ids []UserId) ([]User, error)
	SearchUsers(query string) ([]User, error)
	UpdateDisplayName(name string, id UserId) repositories.MutationWorkItem
	UpdateAvatarId(avatarId *AvatarId, id UserId) repositories.MutationWorkItem
}
