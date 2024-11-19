package defaultVerificationHandler

import (
	"net/http"

	verificationController "github.com/rzmn/governi/internal/controllers/verification"
	"github.com/rzmn/governi/internal/requestHandlers/verification"
	"github.com/rzmn/governi/internal/schema"
	"github.com/rzmn/governi/internal/services/logging"
)

func New(
	controller verificationController.Controller,
	logger logging.Service,
) verification.RequestsHandler {
	return &defaultRequestsHandler{
		controller: controller,
		logger:     logger,
	}
}

type defaultRequestsHandler struct {
	controller verificationController.Controller
	logger     logging.Service
}

func (c *defaultRequestsHandler) ConfirmEmail(
	subject schema.UserId,
	request schema.ConfirmEmailRequest,
	success func(schema.StatusCode, schema.VoidResponse),
	failure func(schema.StatusCode, schema.Response[schema.Error]),
) {
	if err := c.controller.ConfirmEmail(verificationController.UserId(subject), request.Code); err != nil {
		switch err.Code {
		case verificationController.ConfirmEmailErrorWrongConfirmationCode:
			failure(http.StatusConflict, schema.Failure(err, schema.CodeIncorrectCredentials))
		default:
			c.logger.LogError("confirmEmail request %v failed with unknown err: %v", request, err)
			failure(http.StatusInternalServerError, schema.Failure(err, schema.CodeInternal))
		}
	}
	success(http.StatusOK, schema.OK())
}

func (c *defaultRequestsHandler) SendEmailConfirmationCode(
	subject schema.UserId,
	success func(schema.StatusCode, schema.VoidResponse),
	failure func(schema.StatusCode, schema.Response[schema.Error]),
) {
	if err := c.controller.SendConfirmationCode(verificationController.UserId(subject)); err != nil {
		switch err.Code {
		default:
			c.logger.LogError("sendEmailConfirmationCode request failed with unknown err: %v", err)
			failure(http.StatusInternalServerError, schema.Failure(err, schema.CodeInternal))
		}
	}
	success(http.StatusOK, schema.OK())
}
