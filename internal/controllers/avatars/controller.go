package avatars

import (
	"verni/internal/common"
	imagesRepository "verni/internal/repositories/images"
)

type AvatarId string
type Avatar imagesRepository.Image

type Controller interface {
	GetAvatars(ids []AvatarId) ([]Avatar, *common.CodeBasedError[GetAvatarsErrorCode])
}
