package profile

import (
	"log"
	"verni/internal/common"
	"verni/internal/storage"
)

type defaultController struct {
	storage storage.Storage
}

func (c *defaultController) GetProfileInfo(id UserId) (ProfileInfo, *common.CodeBasedError[GetInfoErrorCode]) {
	const op = "profile.defaultController.GetProfileInfo"
	log.Printf("%s: start[id=%s]", op, id)
	info, err := c.storage.GetAccountInfo(storage.UserId(id))
	if err != nil {
		log.Printf("%s: failed read from db err: %v", op, err)
		return ProfileInfo{}, common.NewError(GetInfoErrorInternal)
	}
	if info == nil {
		log.Printf("%s: not found", op)
		return ProfileInfo{}, common.NewError(GetInfoErrorNotFound)
	}
	log.Printf("%s: success[id=%s]", op, id)
	return ProfileInfo(*info), nil
}

func (c *defaultController) UpdateDisplayName(name string, id UserId) *common.CodeBasedError[UpdateDisplayNameErrorCode] {
	const op = "profile.defaultController.UpdateDisplayName"
	log.Printf("%s: start[id=%s name=%s]", op, id, name)
	if err := validateDisplayNameFormat(name); err != nil {
		log.Printf("%s: invalid display name format err: %v", op, err)
		return common.NewError(UpdateDisplayNameErrorWrongFormat)
	}
	if err := c.storage.StoreDisplayName(storage.UserId(id), name); err != nil {
		log.Printf("%s: cannot write to db err: %v", op, err)
		return common.NewError(UpdateDisplayNameErrorInternal)
	}
	log.Printf("%s: success[id=%s name=%s]", op, id, name)
	return nil
}

func (c *defaultController) UpdateAvatar(base64 string, id UserId) (AvatarId, *common.CodeBasedError[UpdateAvatarErrorCode]) {
	const op = "profile.defaultController.UpdateAvatar"
	log.Printf("%s: start[id=%s, base64 len=%d]", op, id, len(base64))
	aid, err := c.storage.StoreAvatarBase64(storage.UserId(id), base64)
	if err != nil {
		log.Printf("%s: cannot write to db err: %v", op, err)
		return AvatarId(aid), common.NewError(UpdateAvatarErrorInternal)
	}
	log.Printf("%s: success[id=%s, base64 len=%d]", op, id, len(base64))
	return AvatarId(aid), nil
}
