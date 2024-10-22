package profile

import (
	"errors"
	"log"
	"verni/internal/common"
	"verni/internal/repositories/auth"
	"verni/internal/repositories/users"
)

type defaultController struct {
	auth    AuthRepository
	images  ImagesRepository
	users   UsersRepository
	friends FriendsRepository
}

func (c *defaultController) GetProfileInfo(id UserId) (ProfileInfo, *common.CodeBasedError[GetInfoErrorCode]) {
	const op = "profile.defaultController.GetProfileInfo"
	log.Printf("%s: start[id=%s]", op, id)

	users, err := c.users.GetUsers([]users.UserId{users.UserId(id)})
	if err != nil {
		log.Printf("%s: cannot get user info err: %v", op, err)
		return ProfileInfo{}, common.NewErrorWithDescription(GetInfoErrorInternal, err.Error())
	}
	if len(users) == 0 {
		err := errors.New("no such user exists")
		log.Printf("%s: cannot get user info err: %v", op, err)
		return ProfileInfo{}, common.NewErrorWithDescription(GetInfoErrorInternal, err.Error())
	}
	credentials, err := c.auth.GetUserInfo(auth.UserId(id))
	if err != nil {
		log.Printf("%s: cannot get user credentials err: %v", op, err)
		return ProfileInfo{}, common.NewErrorWithDescription(GetInfoErrorInternal, err.Error())
	}

	log.Printf("%s: success[id=%s]", op, id)
	return ProfileInfo{
		Id:            id,
		DisplayName:   users[0].DisplayName,
		AvatarId:      (*AvatarId)(users[0].AvatarId),
		Email:         credentials.Email,
		EmailVerified: credentials.EmailVerified,
	}, nil
}

func (c *defaultController) UpdateDisplayName(name string, id UserId) *common.CodeBasedError[UpdateDisplayNameErrorCode] {
	const op = "profile.defaultController.UpdateDisplayName"
	log.Printf("%s: start[id=%s name=%s]", op, id, name)
	if err := validateDisplayNameFormat(name); err != nil {
		log.Printf("%s: invalid display name format err: %v", op, err)
		return common.NewError(UpdateDisplayNameErrorWrongFormat)
	}
	transaction := c.users.UpdateDisplayName(name, users.UserId(id))
	if err := transaction.Perform(); err != nil {
		log.Printf("%s: cannot write to db err: %v", op, err)
		return common.NewError(UpdateDisplayNameErrorInternal)
	}
	log.Printf("%s: success[id=%s name=%s]", op, id, name)
	return nil
}

func (c *defaultController) UpdateAvatar(base64 string, id UserId) (AvatarId, *common.CodeBasedError[UpdateAvatarErrorCode]) {
	const op = "profile.defaultController.UpdateAvatar"
	log.Printf("%s: start[id=%s, base64 len=%d]", op, id, len(base64))
	uploadImageTransaction := c.images.UploadImageBase64(base64)
	aid, err := uploadImageTransaction.Perform()
	if err != nil {
		log.Printf("%s: cannot write to db err: %v", op, err)
		return AvatarId(aid), common.NewError(UpdateAvatarErrorInternal)
	}
	updateAvatarTransaction := c.users.UpdateAvatarId((*users.AvatarId)(&aid), users.UserId(id))
	if err := updateAvatarTransaction.Perform(); err != nil {
		uploadImageTransaction.Rollback()
		log.Printf("%s: cannot write to db err: %v", op, err)
		return AvatarId(aid), common.NewError(UpdateAvatarErrorInternal)
	}
	log.Printf("%s: success[id=%s, base64 len=%d]", op, id, len(base64))
	return AvatarId(aid), nil
}
