package avatars

import (
	"verni/internal/common"
	"verni/internal/storage"
)

type AvatarId storage.AvatarId
type AvatarData storage.AvatarData

type Controller interface {
	GetAvatars(ids []AvatarId) (map[AvatarId]AvatarData, *common.CodeBasedError[GetAvatarsErrorCode])
}

func DefaultController(storage storage.Storage) Controller {
	return &defaultController{
		storage: storage,
	}
}
