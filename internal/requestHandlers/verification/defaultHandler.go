package verification

import (
	"net/http"
	"verni/internal/common"
	verificationController "verni/internal/controllers/verification"
	"verni/internal/schema"
	"verni/internal/services/logging"
)

type defaultRequestsHandler struct {
	controller verificationController.Controller
	logger     logging.Service
}

func (c *defaultRequestsHandler) ConfirmEmail(
	subject schema.UserId,
	request ConfirmEmailRequest,
	success func(schema.StatusCode, schema.VoidResponse),
	failure func(schema.StatusCode, schema.Response[schema.Error]),
) {
	if err := c.controller.ConfirmEmail(verificationController.UserId(subject), request.Code); err != nil {
		switch err.Code {
		case verificationController.ConfirmEmailErrorWrongConfirmationCode:
			failure(
				http.StatusConflict,
				schema.Failure(
					common.NewErrorWithDescriptionValue(
						schema.CodeIncorrectCredentials,
						err.Error(),
					),
				),
			)
		default:
			c.logger.LogError("confirmEmail request %v failed with unknown err: %v", request, err)
			failure(
				http.StatusInternalServerError,
				schema.Failure(
					common.NewErrorWithDescriptionValue(
						schema.CodeInternal,
						err.Error(),
					),
				),
			)
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
			failure(
				http.StatusInternalServerError,
				schema.Failure(
					common.NewErrorWithDescriptionValue(
						schema.CodeInternal,
						err.Error(),
					),
				),
			)
		}
	}
	success(http.StatusOK, schema.OK())
}
