package defaultAvatarsHandler

import (
	"net/http"

	"github.com/rzmn/governi/internal/common"
	avatarsController "github.com/rzmn/governi/internal/controllers/avatars"
	"github.com/rzmn/governi/internal/requestHandlers/avatars"
	"github.com/rzmn/governi/internal/schema"
	"github.com/rzmn/governi/internal/services/logging"
)

func New(
	controller avatarsController.Controller,
	logger logging.Service,
) avatars.RequestsHandler {
	return &defaultRequestsHandler{
		controller: controller,
		logger:     logger,
	}
}

type defaultRequestsHandler struct {
	controller avatarsController.Controller
	logger     logging.Service
}

func (c *defaultRequestsHandler) GetAvatars(
	request schema.GetAvatarsRequest,
	success func(schema.StatusCode, schema.Response[map[schema.ImageId]schema.Image]),
	failure func(schema.StatusCode, schema.Response[schema.Error]),
) {
	info, err := c.controller.GetAvatars(common.Map(request.Ids, func(id schema.ImageId) avatarsController.AvatarId {
		return avatarsController.AvatarId(id)
	}))
	if err != nil {
		switch err.Code {
		default:
			c.logger.LogError("getAvatars request %v failed with unknown err: %v", request, err)
			failure(http.StatusInternalServerError, schema.Failure(err, schema.CodeInternal))
		}
		return
	}
	response := map[schema.ImageId]schema.Image{}
	for _, avatar := range info {
		response[schema.ImageId(avatar.Id)] = schema.Image{
			Id:         schema.ImageId(avatar.Id),
			Base64Data: avatar.Base64,
		}
	}
	success(http.StatusOK, schema.Success(response))
}
