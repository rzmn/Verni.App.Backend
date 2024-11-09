package auth

import (
	"net/http"
	"verni/internal/common"
	authController "verni/internal/controllers/auth"
	httpserver "verni/internal/http-server"
	"verni/internal/http-server/responses"
	"verni/internal/services/logging"
)

type defaultRequestsHandler struct {
	controller authController.Controller
	logger     logging.Service
}

func (c *defaultRequestsHandler) Signup(
	request SignupRequest,
	success func(httpserver.StatusCode, responses.Response[httpserver.Session]),
	failure func(httpserver.StatusCode, responses.Response[responses.Error]),
) {
	session, err := c.controller.Signup(request.Credentials.Email, request.Credentials.Password)
	if err != nil {
		switch err.Code {
		case authController.SignupErrorAlreadyTaken:
			failure(
				http.StatusConflict,
				responses.Failure(
					common.NewErrorWithDescriptionValue(
						responses.CodeAlreadyTaken,
						err.Error(),
					),
				),
			)
		case authController.SignupErrorWrongFormat:
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
			c.logger.LogError("signup request %v failed with unknown err: %v", request, err)
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
	success(http.StatusOK, responses.Success(mapSession(session)))
}

func (c *defaultRequestsHandler) Login(
	request LoginRequest,
	success func(httpserver.StatusCode, responses.Response[httpserver.Session]),
	failure func(httpserver.StatusCode, responses.Response[responses.Error]),
) {
	session, err := c.controller.Login(request.Credentials.Email, request.Credentials.Password)
	if err != nil {
		switch err.Code {
		case authController.LoginErrorWrongCredentials:
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
			c.logger.LogError("login request %v failed with unknown err: %v", request, err)
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
	success(http.StatusOK, responses.Success(mapSession(session)))
}

func (c *defaultRequestsHandler) Refresh(
	request RefreshRequest,
	success func(httpserver.StatusCode, responses.Response[httpserver.Session]),
	failure func(httpserver.StatusCode, responses.Response[responses.Error]),
) {
	session, err := c.controller.Refresh(request.RefreshToken)
	if err != nil {
		switch err.Code {
		case authController.RefreshErrorTokenExpired:
			failure(
				http.StatusUnauthorized,
				responses.Failure(
					common.NewErrorWithDescriptionValue(
						responses.CodeTokenExpired,
						err.Error(),
					),
				),
			)
		case authController.RefreshErrorTokenIsWrong:
			failure(
				http.StatusConflict,
				responses.Failure(
					common.NewErrorWithDescriptionValue(
						responses.CodeWrongAccessToken,
						err.Error(),
					),
				),
			)
		default:
			c.logger.LogError("refresh request %v failed with unknown err: %v", request, err)
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
	success(http.StatusOK, responses.Success(mapSession(session)))
}

func (c *defaultRequestsHandler) UpdateEmail(
	subject httpserver.UserId,
	request UpdateEmailRequest,
	success func(httpserver.StatusCode, responses.Response[httpserver.Session]),
	failure func(httpserver.StatusCode, responses.Response[responses.Error]),
) {
	session, err := c.controller.UpdateEmail(request.Email, authController.UserId(subject))
	if err != nil {
		switch err.Code {
		case authController.UpdateEmailErrorAlreadyTaken:
			failure(
				http.StatusConflict,
				responses.Failure(
					common.NewErrorWithDescriptionValue(
						responses.CodeAlreadyTaken,
						err.Error(),
					),
				),
			)
		case authController.UpdateEmailErrorWrongFormat:
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
			c.logger.LogError("updateEmail request %v failed with unknown err: %v", request, err)
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
	success(http.StatusOK, responses.Success(mapSession(session)))
}

func (c *defaultRequestsHandler) UpdatePassword(
	subject httpserver.UserId,
	request UpdatePasswordRequest,
	success func(httpserver.StatusCode, responses.Response[httpserver.Session]),
	failure func(httpserver.StatusCode, responses.Response[responses.Error]),
) {
	session, err := c.controller.UpdatePassword(request.OldPassword, request.NewPassword, authController.UserId(subject))
	if err != nil {
		switch err.Code {
		case authController.UpdatePasswordErrorOldPasswordIsWrong:
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
			c.logger.LogError("updatePassword request %v failed with unknown err: %v", request, err)
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
	success(http.StatusOK, responses.Success(mapSession(session)))
}

func (c *defaultRequestsHandler) RegisterForPushNotifications(
	subject httpserver.UserId,
	request RegisterForPushNotificationsRequest,
	success func(httpserver.StatusCode, responses.VoidResponse),
	failure func(httpserver.StatusCode, responses.Response[responses.Error]),
) {
	err := c.controller.RegisterForPushNotifications(request.Token, authController.UserId(subject))
	if err != nil {
		switch err.Code {
		default:
			c.logger.LogError("registerForPushNotifications request %v failed with unknown err: %v", request, err)
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

func (c *defaultRequestsHandler) Logout(
	subject httpserver.UserId,
	success func(httpserver.StatusCode, responses.VoidResponse),
	failure func(httpserver.StatusCode, responses.Response[responses.Error]),
) {
	err := c.controller.Logout(authController.UserId(subject))
	if err != nil {
		switch err.Code {
		default:
			c.logger.LogError("logout request failed with unknown err: %v", err)
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

func mapSession(session authController.Session) httpserver.Session {
	return httpserver.Session{
		Id:           httpserver.UserId(session.Id),
		AccessToken:  session.AccessToken,
		RefreshToken: session.RefreshToken,
	}
}
