package avatars

import (
	"github.com/rzmn/Verni.App.Backend/internal/common"
	imagesRepository "github.com/rzmn/Verni.App.Backend/internal/repositories/images"
)

type AvatarId string
type Avatar imagesRepository.Image

type Controller interface {
	GetAvatars(ids []AvatarId) ([]Avatar, *common.CodeBasedError[GetAvatarsErrorCode])
}
