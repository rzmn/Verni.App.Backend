package avatars

import (
	"verni/internal/common"
	imagesRepository "verni/internal/repositories/images"
)

type AvatarId string
type Avatar imagesRepository.Image
type Repository imagesRepository.Repository

type Controller interface {
	GetAvatars(ids []AvatarId) ([]Avatar, *common.CodeBasedError[GetAvatarsErrorCode])
}

func DefaultController(repository Repository) Controller {
	return &defaultController{
		repository: repository,
	}
}
