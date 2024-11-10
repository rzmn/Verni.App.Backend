package verification

import (
	verificationController "verni/internal/controllers/verification"
	httpserver "verni/internal/http-server"
	"verni/internal/http-server/responses"
	"verni/internal/services/logging"
)

type ConfirmEmailRequest struct {
	Code string `json:"code"`
}

type RequestsHandler interface {
	ConfirmEmail(
		subject httpserver.UserId,
		request ConfirmEmailRequest,
		success func(httpserver.StatusCode, responses.VoidResponse),
		failure func(httpserver.StatusCode, responses.Response[responses.Error]),
	)
	SendEmailConfirmationCode(
		subject httpserver.UserId,
		success func(httpserver.StatusCode, responses.VoidResponse),
		failure func(httpserver.StatusCode, responses.Response[responses.Error]),
	)
}

func DefaultHandler(
	controller verificationController.Controller,
	logger logging.Service,
) RequestsHandler {
	return &defaultRequestsHandler{
		controller: controller,
		logger:     logger,
	}
}
