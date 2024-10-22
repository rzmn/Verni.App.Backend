package avatars

import (
	"verni/internal/common"
	"verni/internal/storage"
)

type AvatarId string
type Avatar struct {
	Id         AvatarId
	Base64Data *string
}

type Controller interface {
	GetAvatars(ids []AvatarId) (map[AvatarId]Avatar, *common.CodeBasedError[GetAvatarsErrorCode])
}

func DefaultController(storage storage.Storage) Controller {
	return &defaultController{
		storage: storage,
	}
}
