package auth

import (
	authController "verni/internal/controllers/auth"
	httpserver "verni/internal/http-server"
	"verni/internal/services/logging"
)

type SignupRequest struct {
	Credentials httpserver.Credentials `json:"credentials"`
}

type LoginRequest struct {
	Credentials httpserver.Credentials `json:"credentials"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type UpdateEmailRequest struct {
	Email string `json:"email"`
}

type UpdatePasswordRequest struct {
	OldPassword string `json:"old"`
	NewPassword string `json:"new"`
}

type RegisterForPushNotificationsRequest struct {
	Token string `json:"token"`
}

type RequestsHandler interface {
	Signup(
		request SignupRequest,
		success func(httpserver.StatusCode, httpserver.Response[httpserver.Session]),
		failure func(httpserver.StatusCode, httpserver.Response[httpserver.Error]),
	)
	Login(
		request LoginRequest,
		success func(httpserver.StatusCode, httpserver.Response[httpserver.Session]),
		failure func(httpserver.StatusCode, httpserver.Response[httpserver.Error]),
	)
	Refresh(
		request RefreshRequest,
		success func(httpserver.StatusCode, httpserver.Response[httpserver.Session]),
		failure func(httpserver.StatusCode, httpserver.Response[httpserver.Error]),
	)
	UpdateEmail(
		subject httpserver.UserId,
		request UpdateEmailRequest,
		success func(httpserver.StatusCode, httpserver.Response[httpserver.Session]),
		failure func(httpserver.StatusCode, httpserver.Response[httpserver.Error]),
	)
	UpdatePassword(
		subject httpserver.UserId,
		request UpdatePasswordRequest,
		success func(httpserver.StatusCode, httpserver.Response[httpserver.Session]),
		failure func(httpserver.StatusCode, httpserver.Response[httpserver.Error]),
	)
	RegisterForPushNotifications(
		subject httpserver.UserId,
		request RegisterForPushNotificationsRequest,
		success func(httpserver.StatusCode, httpserver.VoidResponse),
		failure func(httpserver.StatusCode, httpserver.Response[httpserver.Error]),
	)
	Logout(
		subject httpserver.UserId,
		success func(httpserver.StatusCode, httpserver.VoidResponse),
		failure func(httpserver.StatusCode, httpserver.Response[httpserver.Error]),
	)
}

func DefaultHandler(
	controller authController.Controller,
	logger logging.Service,
) RequestsHandler {
	return &defaultRequestsHandler{
		controller: controller,
		logger:     logger,
	}
}
