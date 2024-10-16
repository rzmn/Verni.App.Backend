package profile

import (
	"verni/internal/common"
	"verni/internal/storage"
)

type UserId storage.UserId
type ProfileInfo storage.ProfileInfo
type AvatarId storage.AvatarId

type Controller interface {
	GetProfileInfo(id UserId) (ProfileInfo, *common.CodeBasedError[GetInfoErrorCode])
	UpdateDisplayName(name string, id UserId) *common.CodeBasedError[UpdateDisplayNameErrorCode]
	UpdateAvatar(base64 string, id UserId) (AvatarId, *common.CodeBasedError[UpdateAvatarErrorCode])
}

func DefaultController(storage storage.Storage) Controller {
	return &defaultController{
		storage: storage,
	}
}
