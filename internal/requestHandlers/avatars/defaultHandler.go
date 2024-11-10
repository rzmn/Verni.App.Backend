package avatars

import (
	"net/http"
	"verni/internal/common"
	avatarsController "verni/internal/controllers/avatars"
	httpserver "verni/internal/http-server"
	"verni/internal/services/logging"
)

type defaultRequestsHandler struct {
	controller avatarsController.Controller
	logger     logging.Service
}

func (c *defaultRequestsHandler) GetAvatars(
	request GetAvatarsRequest,
	success func(httpserver.StatusCode, httpserver.Response[map[httpserver.ImageId]httpserver.Image]),
	failure func(httpserver.StatusCode, httpserver.Response[httpserver.Error]),
) {
	info, err := c.controller.GetAvatars(common.Map(request.Ids, func(id httpserver.ImageId) avatarsController.AvatarId {
		return avatarsController.AvatarId(id)
	}))
	if err != nil {
		switch err.Code {
		default:
			c.logger.LogError("getAvatars request %v failed with unknown err: %v", request, err)
			failure(
				http.StatusInternalServerError,
				httpserver.Failure(
					common.NewErrorWithDescriptionValue(
						httpserver.CodeInternal,
						err.Error(),
					),
				),
			)
		}
		return
	}
	response := map[httpserver.ImageId]httpserver.Image{}
	for _, avatar := range info {
		response[httpserver.ImageId(avatar.Id)] = httpserver.Image{
			Id:         httpserver.ImageId(avatar.Id),
			Base64Data: avatar.Base64,
		}
	}
	success(http.StatusOK, httpserver.Success(response))
}
