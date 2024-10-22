package profile

import (
	"verni/internal/common"
	authRepository "verni/internal/repositories/auth"
	friendsRepository "verni/internal/repositories/friends"
	imagesRepository "verni/internal/repositories/images"
	usersRepository "verni/internal/repositories/users"
)

type UserId string
type AvatarId string

type ProfileInfo struct {
	Id            UserId
	DisplayName   string
	AvatarId      *AvatarId
	Email         string
	EmailVerified bool
}

type AuthRepository authRepository.Repository
type ImagesRepository imagesRepository.Repository
type UsersRepository usersRepository.Repository
type FriendsRepository friendsRepository.Repository

type Controller interface {
	GetProfileInfo(id UserId) (ProfileInfo, *common.CodeBasedError[GetInfoErrorCode])
	UpdateDisplayName(name string, id UserId) *common.CodeBasedError[UpdateDisplayNameErrorCode]
	UpdateAvatar(base64 string, id UserId) (AvatarId, *common.CodeBasedError[UpdateAvatarErrorCode])
}

func DefaultController(
	auth AuthRepository,
	images ImagesRepository,
	users UsersRepository,
	friends FriendsRepository,
) Controller {
	return &defaultController{
		auth:    auth,
		images:  images,
		users:   users,
		friends: friends,
	}
}
