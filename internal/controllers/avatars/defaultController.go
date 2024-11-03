package avatars

import (
	"verni/internal/common"
	imagesRepository "verni/internal/repositories/images"
	"verni/internal/services/logging"
)

type defaultController struct {
	repository Repository
	logger     logging.Service
}

func (c *defaultController) GetAvatars(ids []AvatarId) ([]Avatar, *common.CodeBasedError[GetAvatarsErrorCode]) {
	const op = "avatars.defaultController.GetAvatars"
	c.logger.Log("%s: start[ids=%s]", op, ids)
	avatars, err := c.repository.GetImagesBase64(common.Map(ids, func(id AvatarId) imagesRepository.ImageId {
		return imagesRepository.ImageId(id)
	}))
	if err != nil {
		c.logger.Log("%s: cannot read from db %v", op, err)
		return []Avatar{}, common.NewError(GetAvatarsErrorInternal)
	}
	c.logger.Log("%s: success[ids=%s]", op, ids)
	return common.Map(avatars, func(image imagesRepository.Image) Avatar {
		return Avatar(image)
	}), nil
}
