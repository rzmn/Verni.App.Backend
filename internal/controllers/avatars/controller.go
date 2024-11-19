package avatars

import (
	"github.com/rzmn/governi/internal/common"
	imagesRepository "github.com/rzmn/governi/internal/repositories/images"
)

type AvatarId string
type Avatar imagesRepository.Image

type Controller interface {
	GetAvatars(ids []AvatarId) ([]Avatar, *common.CodeBasedError[GetAvatarsErrorCode])
}
