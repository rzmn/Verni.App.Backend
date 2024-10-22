package avatars

import (
	"log"
	"verni/internal/common"
	imagesRepository "verni/internal/repositories/images"
)

type defaultController struct {
	repository Repository
}

func (c *defaultController) GetAvatars(ids []AvatarId) ([]Avatar, *common.CodeBasedError[GetAvatarsErrorCode]) {
	const op = "avatars.defaultController.GetAvatars"
	log.Printf("%s: start[ids=%s]", op, ids)
	avatars, err := c.repository.GetImagesBase64(common.Map(ids, func(id AvatarId) imagesRepository.ImageId {
		return imagesRepository.ImageId(id)
	}))
	if err != nil {
		log.Printf("%s: cannot read from db %v", op, err)
		return []Avatar{}, common.NewError(GetAvatarsErrorInternal)
	}
	log.Printf("%s: success[ids=%s]", op, ids)
	return common.Map(avatars, func(image imagesRepository.Image) Avatar {
		return Avatar(image)
	}), nil
}
