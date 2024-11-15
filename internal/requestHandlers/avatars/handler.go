package avatars

import (
	avatarsController "verni/internal/controllers/avatars"
	"verni/internal/schema"
	"verni/internal/services/logging"
)

type RequestsHandler interface {
	GetAvatars(
		request schema.GetAvatarsRequest,
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
