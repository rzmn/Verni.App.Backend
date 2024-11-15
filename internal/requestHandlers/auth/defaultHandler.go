package auth

import (
	"net/http"
	authController "verni/internal/controllers/auth"
	"verni/internal/schema"
	"verni/internal/services/logging"
)

type defaultRequestsHandler struct {
	controller authController.Controller
	logger     logging.Service
}

func (c *defaultRequestsHandler) Signup(
	request schema.SignupRequest,
	success func(schema.StatusCode, schema.Response[schema.Session]),
	failure func(schema.StatusCode, schema.Response[schema.Error]),
) {
	session, err := c.controller.Signup(request.Credentials.Email, request.Credentials.Password)
	if err != nil {
		switch err.Code {
		case authController.SignupErrorAlreadyTaken:
			failure(http.StatusConflict, schema.Failure(err, schema.CodeAlreadyTaken))
		case authController.SignupErrorWrongFormat:
			failure(http.StatusUnprocessableEntity, schema.Failure(err, schema.CodeWrongFormat))
		default:
			c.logger.LogError("signup request %v failed with unknown err: %v", request, err)
			failure(http.StatusInternalServerError, schema.Failure(err, schema.CodeInternal))
		}
		return
	}
	success(http.StatusOK, schema.Success(mapSession(session)))
}

func (c *defaultRequestsHandler) Login(
	request schema.LoginRequest,
	success func(schema.StatusCode, schema.Response[schema.Session]),
	failure func(schema.StatusCode, schema.Response[schema.Error]),
) {
	session, err := c.controller.Login(request.Credentials.Email, request.Credentials.Password)
	if err != nil {
		switch err.Code {
		case authController.LoginErrorWrongCredentials:
			failure(http.StatusConflict, schema.Failure(err, schema.CodeIncorrectCredentials))
		default:
			c.logger.LogError("login request %v failed with unknown err: %v", request, err)
			failure(http.StatusInternalServerError, schema.Failure(err, schema.CodeInternal))
		}
		return
	}
	success(http.StatusOK, schema.Success(mapSession(session)))
}

func (c *defaultRequestsHandler) Refresh(
	request schema.RefreshRequest,
	success func(schema.StatusCode, schema.Response[schema.Session]),
	failure func(schema.StatusCode, schema.Response[schema.Error]),
) {
	session, err := c.controller.Refresh(request.RefreshToken)
	if err != nil {
		switch err.Code {
		case authController.RefreshErrorTokenExpired:
			failure(http.StatusUnauthorized, schema.Failure(err, schema.CodeTokenExpired))
		case authController.RefreshErrorTokenIsWrong:
			failure(http.StatusConflict, schema.Failure(err, schema.CodeWrongAccessToken))
		default:
			c.logger.LogError("refresh request %v failed with unknown err: %v", request, err)
			failure(http.StatusInternalServerError, schema.Failure(err, schema.CodeInternal))
		}
		return
	}
	success(http.StatusOK, schema.Success(mapSession(session)))
}

func (c *defaultRequestsHandler) UpdateEmail(
	subject schema.UserId,
	request schema.UpdateEmailRequest,
	success func(schema.StatusCode, schema.Response[schema.Session]),
	failure func(schema.StatusCode, schema.Response[schema.Error]),
) {
	session, err := c.controller.UpdateEmail(request.Email, authController.UserId(subject))
	if err != nil {
		switch err.Code {
		case authController.UpdateEmailErrorAlreadyTaken:
			failure(http.StatusConflict, schema.Failure(err, schema.CodeAlreadyTaken))
		case authController.UpdateEmailErrorWrongFormat:
			failure(http.StatusUnprocessableEntity, schema.Failure(err, schema.CodeWrongFormat))
		default:
			c.logger.LogError("updateEmail request %v failed with unknown err: %v", request, err)
			failure(http.StatusInternalServerError, schema.Failure(err, schema.CodeInternal))
		}
		return
	}
	success(http.StatusOK, schema.Success(mapSession(session)))
}

func (c *defaultRequestsHandler) UpdatePassword(
	subject schema.UserId,
	request schema.UpdatePasswordRequest,
	success func(schema.StatusCode, schema.Response[schema.Session]),
	failure func(schema.StatusCode, schema.Response[schema.Error]),
) {
	session, err := c.controller.UpdatePassword(request.OldPassword, request.NewPassword, authController.UserId(subject))
	if err != nil {
		switch err.Code {
		case authController.UpdatePasswordErrorOldPasswordIsWrong:
			failure(http.StatusConflict, schema.Failure(err, schema.CodeIncorrectCredentials))
		default:
			c.logger.LogError("updatePassword request %v failed with unknown err: %v", request, err)
			failure(http.StatusInternalServerError, schema.Failure(err, schema.CodeInternal))
		}
		return
	}
	success(http.StatusOK, schema.Success(mapSession(session)))
}

func (c *defaultRequestsHandler) RegisterForPushNotifications(
	subject schema.UserId,
	request schema.RegisterForPushNotificationsRequest,
	success func(schema.StatusCode, schema.VoidResponse),
	failure func(schema.StatusCode, schema.Response[schema.Error]),
) {
	err := c.controller.RegisterForPushNotifications(request.Token, authController.UserId(subject))
	if err != nil {
		switch err.Code {
		default:
			c.logger.LogError("registerForPushNotifications request %v failed with unknown err: %v", request, err)
			failure(http.StatusInternalServerError, schema.Failure(err, schema.CodeInternal))
		}
		return
	}
	success(http.StatusOK, schema.OK())
}

func (c *defaultRequestsHandler) Logout(
	subject schema.UserId,
	success func(schema.StatusCode, schema.VoidResponse),
	failure func(schema.StatusCode, schema.Response[schema.Error]),
) {
	err := c.controller.Logout(authController.UserId(subject))
	if err != nil {
		switch err.Code {
		default:
			c.logger.LogError("logout request failed with unknown err: %v", err)
			failure(http.StatusInternalServerError, schema.Failure(err, schema.CodeInternal))
		}
		return
	}
	success(http.StatusOK, schema.OK())
}

func mapSession(session authController.Session) schema.Session {
	return schema.Session{
		Id:           schema.UserId(session.Id),
		AccessToken:  session.AccessToken,
		RefreshToken: session.RefreshToken,
	}
}
