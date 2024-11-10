package profile

import (
	"net/http"
	"verni/internal/common"
	profileController "verni/internal/controllers/profile"
	httpserver "verni/internal/http-server"
	"verni/internal/http-server/responses"
	"verni/internal/services/logging"
)

type defaultRequestsHandler struct {
	controller profileController.Controller
	logger     logging.Service
}

func (c *defaultRequestsHandler) GetInfo(
	subject httpserver.UserId,
	success func(httpserver.StatusCode, responses.Response[httpserver.Profile]),
	failure func(httpserver.StatusCode, responses.Response[responses.Error]),
) {
	info, err := c.controller.GetProfileInfo(profileController.UserId(subject))
	if err != nil {
		switch err.Code {
		case profileController.GetInfoErrorNotFound:
			failure(
				http.StatusConflict,
				responses.Failure(
					common.NewErrorWithDescriptionValue(
						responses.CodeNoSuchUser,
						err.Error(),
					),
				),
			)
		default:
			c.logger.LogError("getProfile request failed with unknown err: %v", err)
			failure(
				http.StatusInternalServerError,
				responses.Failure(
					common.NewErrorWithDescriptionValue(
						responses.CodeInternal,
						err.Error(),
					),
				),
			)
		}
		return
	}
	success(http.StatusOK, responses.Success(mapProfile(info)))
}

func (c *defaultRequestsHandler) SetAvatar(
	subject httpserver.UserId,
	request SetAvatarRequest,
	success func(httpserver.StatusCode, responses.Response[httpserver.ImageId]),
	failure func(httpserver.StatusCode, responses.Response[responses.Error]),
) {
	aid, err := c.controller.UpdateAvatar(request.DataBase64, profileController.UserId(subject))
	if err != nil {
		switch err.Code {
		default:
			c.logger.LogError("setAvatar request %v failed with unknown err: %v", request, err)
			failure(
				http.StatusInternalServerError,
				responses.Failure(
					common.NewErrorWithDescriptionValue(
						responses.CodeInternal,
						err.Error(),
					),
				),
			)
		}
		return
	}
	success(http.StatusOK, responses.Success(httpserver.ImageId(aid)))
}

func (c *defaultRequestsHandler) SetDisplayName(
	subject httpserver.UserId,
	request SetDisplayNameRequest,
	success func(httpserver.StatusCode, responses.VoidResponse),
	failure func(httpserver.StatusCode, responses.Response[responses.Error]),
) {
	if err := c.controller.UpdateDisplayName(request.DisplayName, profileController.UserId(subject)); err != nil {
		switch err.Code {
		case profileController.UpdateDisplayNameErrorNotFound:
			failure(
				http.StatusConflict,
				responses.Failure(
					common.NewErrorWithDescriptionValue(
						responses.CodeNoSuchUser,
						err.Error(),
					),
				),
			)
		case profileController.UpdateDisplayNameErrorWrongFormat:
			failure(
				http.StatusUnprocessableEntity,
				responses.Failure(
					common.NewErrorWithDescriptionValue(
						responses.CodeWrongFormat,
						err.Error(),
					),
				),
			)
		default:
			c.logger.LogError("setDisplayName request %v failed with unknown err: %v", request, err)
			failure(
				http.StatusInternalServerError,
				responses.Failure(
					common.NewErrorWithDescriptionValue(
						responses.CodeInternal,
						err.Error(),
					),
				),
			)
		}
		return
	}
	success(http.StatusOK, responses.OK())
}

func mapProfile(profile profileController.ProfileInfo) httpserver.Profile {
	return httpserver.Profile{
		User: httpserver.User{
			Id:           httpserver.UserId(profile.Id),
			DisplayName:  profile.DisplayName,
			AvatarId:     (*httpserver.ImageId)(profile.AvatarId),
			FriendStatus: httpserver.FriendStatusMe,
		},
		Email:         profile.Email,
		EmailVerified: profile.EmailVerified,
	}
}
