package defaultController

import (
	"verni/internal/common"
	"verni/internal/controllers/avatars"
	imagesRepository "verni/internal/repositories/images"
	"verni/internal/services/logging"
)

type Repository imagesRepository.Repository

func New(repository Repository, logger logging.Service) avatars.Controller {
	return &defaultController{
		repository: repository,
		logger:     logger,
	}
}

type defaultController struct {
	repository imagesRepository.Repository
	logger     logging.Service
}

func (c *defaultController) GetAvatars(ids []avatars.AvatarId) ([]avatars.Avatar, *common.CodeBasedError[avatars.GetAvatarsErrorCode]) {
	const op = "avatars.defaultController.GetAvatars"
	c.logger.LogInfo("%s: start[ids=%s]", op, ids)
	result, err := c.repository.GetImagesBase64(common.Map(ids, func(id avatars.AvatarId) imagesRepository.ImageId {
		return imagesRepository.ImageId(id)
	}))
	if err != nil {
		c.logger.LogInfo("%s: cannot read from db %v", op, err)
		return []avatars.Avatar{}, common.NewError(avatars.GetAvatarsErrorInternal)
	}
	c.logger.LogInfo("%s: success[ids=%s]", op, ids)
	return common.Map(result, func(image imagesRepository.Image) avatars.Avatar {
		return avatars.Avatar(image)
	}), nil
}
