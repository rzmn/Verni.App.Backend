package avatars

import (
	avatarsController "verni/internal/controllers/avatars"
	"verni/internal/schema"
	"verni/internal/services/logging"
)

type GetAvatarsRequest struct {
	Ids []schema.ImageId `json:"ids"`
}

type RequestsHandler interface {
	GetAvatars(
		request GetAvatarsRequest,
		success func(schema.StatusCode, schema.Response[map[schema.ImageId]schema.Image]),
		failure func(schema.StatusCode, schema.Response[schema.Error]),
	)
}

func DefaultHandler(
	controller avatarsController.Controller,
	logger logging.Service,
) RequestsHandler {
	return &defaultRequestsHandler{
		controller: controller,
		logger:     logger,
	}
}
