package avatars

import (
	"verni/internal/common"
	imagesRepository "verni/internal/repositories/images"
	"verni/internal/services/logging"
)

type AvatarId string
type Avatar imagesRepository.Image
type Repository imagesRepository.Repository

type Controller interface {
	GetAvatars(ids []AvatarId) ([]Avatar, *common.CodeBasedError[GetAvatarsErrorCode])
}

func DefaultController(repository Repository, logger logging.Service) Controller {
	return &defaultController{
		repository: repository,
		logger:     logger,
	}
}
