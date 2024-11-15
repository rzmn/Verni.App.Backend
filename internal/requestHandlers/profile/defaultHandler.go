package profile

import (
	"net/http"
	profileController "verni/internal/controllers/profile"
	"verni/internal/schema"
	"verni/internal/services/logging"
)

type defaultRequestsHandler struct {
	controller profileController.Controller
	logger     logging.Service
}

func (c *defaultRequestsHandler) GetInfo(
	subject schema.UserId,
	success func(schema.StatusCode, schema.Response[schema.Profile]),
	failure func(schema.StatusCode, schema.Response[schema.Error]),
) {
	info, err := c.controller.GetProfileInfo(profileController.UserId(subject))
	if err != nil {
		switch err.Code {
		case profileController.GetInfoErrorNotFound:
			failure(http.StatusConflict, schema.Failure(err, schema.CodeNoSuchUser))
		default:
			c.logger.LogError("getProfile request failed with unknown err: %v", err)
			failure(http.StatusInternalServerError, schema.Failure(err, schema.CodeInternal))
		}
		return
	}
	success(http.StatusOK, schema.Success(mapProfile(info)))
}

func (c *defaultRequestsHandler) SetAvatar(
	subject schema.UserId,
	request schema.SetAvatarRequest,
	success func(schema.StatusCode, schema.Response[schema.ImageId]),
	failure func(schema.StatusCode, schema.Response[schema.Error]),
) {
	aid, err := c.controller.UpdateAvatar(request.DataBase64, profileController.UserId(subject))
	if err != nil {
		switch err.Code {
		default:
			c.logger.LogError("setAvatar request %v failed with unknown err: %v", request, err)
			failure(http.StatusInternalServerError, schema.Failure(err, schema.CodeInternal))
		}
		return
	}
	success(http.StatusOK, schema.Success(schema.ImageId(aid)))
}

func (c *defaultRequestsHandler) SetDisplayName(
	subject schema.UserId,
	request schema.SetDisplayNameRequest,
	success func(schema.StatusCode, schema.VoidResponse),
	failure func(schema.StatusCode, schema.Response[schema.Error]),
) {
	if err := c.controller.UpdateDisplayName(request.DisplayName, profileController.UserId(subject)); err != nil {
		switch err.Code {
		case profileController.UpdateDisplayNameErrorNotFound:
			failure(http.StatusConflict, schema.Failure(err, schema.CodeNoSuchUser))
		case profileController.UpdateDisplayNameErrorWrongFormat:
			failure(http.StatusUnprocessableEntity, schema.Failure(err, schema.CodeWrongFormat))
		default:
			c.logger.LogError("setDisplayName request %v failed with unknown err: %v", request, err)
			failure(http.StatusInternalServerError, schema.Failure(err, schema.CodeInternal))
		}
		return
	}
	success(http.StatusOK, schema.OK())
}

func mapProfile(profile profileController.ProfileInfo) schema.Profile {
	return schema.Profile{
		User: schema.User{
			Id:           schema.UserId(profile.Id),
			DisplayName:  profile.DisplayName,
			AvatarId:     (*schema.ImageId)(profile.AvatarId),
			FriendStatus: schema.FriendStatusMe,
		},
		Email:         profile.Email,
		EmailVerified: profile.EmailVerified,
	}
}
