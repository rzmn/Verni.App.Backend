package auth

import (
	httpserver "verni/internal/http-server"
	"verni/internal/http-server/responses"
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
		success func(httpserver.StatusCode, responses.Response[httpserver.Session]),
		failure func(httpserver.StatusCode, responses.Response[responses.Error]),
	)
	Login(
		request LoginRequest,
		success func(httpserver.StatusCode, responses.Response[httpserver.Session]),
		failure func(httpserver.StatusCode, responses.Response[responses.Error]),
	)
	Refresh(
		request RefreshRequest,
		success func(httpserver.StatusCode, responses.Response[httpserver.Session]),
		failure func(httpserver.StatusCode, responses.Response[responses.Error]),
	)
	UpdateEmail(
		subject httpserver.UserId,
		request UpdateEmailRequest,
		success func(httpserver.StatusCode, responses.Response[httpserver.Session]),
		failure func(httpserver.StatusCode, responses.Response[responses.Error]),
	)
	UpdatePassword(
		subject httpserver.UserId,
		request UpdatePasswordRequest,
		success func(httpserver.StatusCode, responses.Response[httpserver.Session]),
		failure func(httpserver.StatusCode, responses.Response[responses.Error]),
	)
	RegisterForPushNotifications(
		subject httpserver.UserId,
		request RegisterForPushNotificationsRequest,
		success func(httpserver.StatusCode, responses.VoidResponse),
		failure func(httpserver.StatusCode, responses.Response[responses.Error]),
	)
	Logout(
		subject httpserver.UserId,
		success func(httpserver.StatusCode, responses.VoidResponse),
		failure func(httpserver.StatusCode, responses.Response[responses.Error]),
	)
}

func DefaultHandler() RequestsHandler {
	return &defaultRequestsHandler{}
}
