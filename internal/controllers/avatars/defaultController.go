package avatars

import (
	"log"
	"verni/internal/common"
	"verni/internal/storage"
)

type defaultController struct {
	storage storage.Storage
}

func (c *defaultController) GetAvatars(ids []AvatarId) (map[AvatarId]AvatarData, *common.CodeBasedError[GetAvatarsErrorCode]) {
	const op = "avatars.defaultController.GetAvatars"
	log.Printf("%s: start[ids=%s]", op, ids)
	avatars, err := c.storage.GetAvatarsBase64(common.Map(ids, func(id AvatarId) storage.AvatarId {
		return storage.AvatarId(id)
	}))
	if err != nil {
		log.Printf("%s: cannot read from db %v", op, err)
		return map[AvatarId]AvatarData{}, common.NewError(GetAvatarsErrorInternal)
	}
	storageAvatarsData := map[AvatarId]AvatarData{}
	for id, data := range avatars {
		storageAvatarsData[AvatarId(id)] = AvatarData(data)
	}
	log.Printf("%s: success[ids=%s]", op, ids)
	return storageAvatarsData, nil
}
