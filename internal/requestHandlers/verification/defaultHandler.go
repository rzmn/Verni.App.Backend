package verification

import (
	"net/http"
	"verni/internal/common"
	verificationController "verni/internal/controllers/verification"
	httpserver "verni/internal/http-server"
	"verni/internal/http-server/responses"
	"verni/internal/services/logging"
)

type defaultRequestsHandler struct {
	controller verificationController.Controller
	logger     logging.Service
}

func (c *defaultRequestsHandler) ConfirmEmail(
	subject httpserver.UserId,
	request ConfirmEmailRequest,
	success func(httpserver.StatusCode, responses.VoidResponse),
	failure func(httpserver.StatusCode, responses.Response[responses.Error]),
) {
	if err := c.controller.ConfirmEmail(verificationController.UserId(subject), request.Code); err != nil {
		switch err.Code {
		case verificationController.ConfirmEmailErrorWrongConfirmationCode:
			failure(
				http.StatusConflict,
				responses.Failure(
					common.NewErrorWithDescriptionValue(
						responses.CodeIncorrectCredentials,
						err.Error(),
					),
				),
			)
		default:
			c.logger.LogError("confirmEmail request %v failed with unknown err: %v", request, err)
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
	}
	success(http.StatusOK, responses.OK())
}

func (c *defaultRequestsHandler) SendEmailConfirmationCode(
	subject httpserver.UserId,
	success func(httpserver.StatusCode, responses.VoidResponse),
	failure func(httpserver.StatusCode, responses.Response[responses.Error]),
) {
	if err := c.controller.SendConfirmationCode(verificationController.UserId(subject)); err != nil {
		switch err.Code {
		default:
			c.logger.LogError("sendEmailConfirmationCode request failed with unknown err: %v", err)
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
	}
	success(http.StatusOK, responses.OK())
}
