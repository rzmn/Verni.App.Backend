package defaultController

import (
	"errors"
	"verni/internal/common"
	"verni/internal/controllers/profile"
	"verni/internal/repositories/auth"
	"verni/internal/repositories/users"
	"verni/internal/services/formatValidation"
	"verni/internal/services/logging"

	authRepository "verni/internal/repositories/auth"
	friendsRepository "verni/internal/repositories/friends"
	imagesRepository "verni/internal/repositories/images"
	usersRepository "verni/internal/repositories/users"
)

type AuthRepository authRepository.Repository
type ImagesRepository imagesRepository.Repository
type UsersRepository usersRepository.Repository
type FriendsRepository friendsRepository.Repository

func New(
	auth AuthRepository,
	images ImagesRepository,
	users UsersRepository,
	friends FriendsRepository,
	formatValidation formatValidation.Service,
	logger logging.Service,
) profile.Controller {
	return &defaultController{
		auth:             auth,
		images:           images,
		users:            users,
		friends:          friends,
		formatValidation: formatValidation,
		logger:           logger,
	}
}

type defaultController struct {
	auth             AuthRepository
	images           ImagesRepository
	users            UsersRepository
	friends          FriendsRepository
	formatValidation formatValidation.Service
	logger           logging.Service
}

func (c *defaultController) GetProfileInfo(id profile.UserId) (profile.ProfileInfo, *common.CodeBasedError[profile.GetInfoErrorCode]) {
	const op = "profile.defaultController.GetProfileInfo"
	c.logger.LogInfo("%s: start[id=%s]", op, id)

	users, err := c.users.GetUsers([]users.UserId{users.UserId(id)})
	if err != nil {
		c.logger.LogInfo("%s: cannot get user info err: %v", op, err)
		return profile.ProfileInfo{}, common.NewErrorWithDescription(profile.GetInfoErrorInternal, err.Error())
	}
	if len(users) == 0 {
		err := errors.New("no such user exists")
		c.logger.LogInfo("%s: cannot get user info err: %v", op, err)
		return profile.ProfileInfo{}, common.NewErrorWithDescription(profile.GetInfoErrorInternal, err.Error())
	}
	credentials, err := c.auth.GetUserInfo(auth.UserId(id))
	if err != nil {
		c.logger.LogInfo("%s: cannot get user credentials err: %v", op, err)
		return profile.ProfileInfo{}, common.NewErrorWithDescription(profile.GetInfoErrorInternal, err.Error())
	}

	c.logger.LogInfo("%s: success[id=%s]", op, id)
	return profile.ProfileInfo{
		Id:            id,
		DisplayName:   users[0].DisplayName,
		AvatarId:      (*profile.AvatarId)(users[0].AvatarId),
		Email:         credentials.Email,
		EmailVerified: credentials.EmailVerified,
	}, nil
}

func (c *defaultController) UpdateDisplayName(name string, id profile.UserId) *common.CodeBasedError[profile.UpdateDisplayNameErrorCode] {
	const op = "profile.defaultController.UpdateDisplayName"
	c.logger.LogInfo("%s: start[id=%s name=%s]", op, id, name)
	if err := c.formatValidation.ValidateDisplayNameFormat(name); err != nil {
		c.logger.LogInfo("%s: invalid display name format err: %v", op, err)
		return common.NewError(profile.UpdateDisplayNameErrorWrongFormat)
	}
	transaction := c.users.UpdateDisplayName(name, users.UserId(id))
	if err := transaction.Perform(); err != nil {
		c.logger.LogInfo("%s: cannot write to db err: %v", op, err)
		return common.NewError(profile.UpdateDisplayNameErrorInternal)
	}
	c.logger.LogInfo("%s: success[id=%s name=%s]", op, id, name)
	return nil
}

func (c *defaultController) UpdateAvatar(base64 string, id profile.UserId) (profile.AvatarId, *common.CodeBasedError[profile.UpdateAvatarErrorCode]) {
	const op = "profile.defaultController.UpdateAvatar"
	c.logger.LogInfo("%s: start[id=%s, base64 len=%d]", op, id, len(base64))
	uploadImageTransaction := c.images.UploadImageBase64(base64)
	aid, err := uploadImageTransaction.Perform()
	if err != nil {
		c.logger.LogInfo("%s: cannot write to db err: %v", op, err)
		return profile.AvatarId(aid), common.NewError(profile.UpdateAvatarErrorInternal)
	}
	updateAvatarTransaction := c.users.UpdateAvatarId((*users.AvatarId)(&aid), users.UserId(id))
	if err := updateAvatarTransaction.Perform(); err != nil {
		uploadImageTransaction.Rollback()
		c.logger.LogInfo("%s: cannot write to db err: %v", op, err)
		return profile.AvatarId(aid), common.NewError(profile.UpdateAvatarErrorInternal)
	}
	c.logger.LogInfo("%s: success[id=%s, base64 len=%d]", op, id, len(base64))
	return profile.AvatarId(aid), nil
}
