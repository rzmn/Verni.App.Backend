package verification

import (
	"net/http"
	"verni/internal/common"
	verificationController "verni/internal/controllers/verification"
	httpserver "verni/internal/http-server"
	"verni/internal/services/logging"
)

type defaultRequestsHandler struct {
	controller verificationController.Controller
	logger     logging.Service
}

func (c *defaultRequestsHandler) ConfirmEmail(
	subject httpserver.UserId,
	request ConfirmEmailRequest,
	success func(httpserver.StatusCode, httpserver.VoidResponse),
	failure func(httpserver.StatusCode, httpserver.Response[httpserver.Error]),
) {
	if err := c.controller.ConfirmEmail(verificationController.UserId(subject), request.Code); err != nil {
		switch err.Code {
		case verificationController.ConfirmEmailErrorWrongConfirmationCode:
			failure(
				http.StatusConflict,
				httpserver.Failure(
					common.NewErrorWithDescriptionValue(
						httpserver.CodeIncorrectCredentials,
						err.Error(),
					),
				),
			)
		default:
			c.logger.LogError("confirmEmail request %v failed with unknown err: %v", request, err)
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
	}
	success(http.StatusOK, httpserver.OK())
}

func (c *defaultRequestsHandler) SendEmailConfirmationCode(
	subject httpserver.UserId,
	success func(httpserver.StatusCode, httpserver.VoidResponse),
	failure func(httpserver.StatusCode, httpserver.Response[httpserver.Error]),
) {
	if err := c.controller.SendConfirmationCode(verificationController.UserId(subject)); err != nil {
		switch err.Code {
		default:
			c.logger.LogError("sendEmailConfirmationCode request failed with unknown err: %v", err)
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
	}
	success(http.StatusOK, httpserver.OK())
}
