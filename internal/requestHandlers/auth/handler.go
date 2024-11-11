package auth

import (
	authController "verni/internal/controllers/auth"
	"verni/internal/schema"
	"verni/internal/services/logging"
)

type SignupRequest struct {
	Credentials schema.Credentials `json:"credentials"`
}

type LoginRequest struct {
	Credentials schema.Credentials `json:"credentials"`
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
		success func(schema.StatusCode, schema.Response[schema.Session]),
		failure func(schema.StatusCode, schema.Response[schema.Error]),
	)
	Login(
		request LoginRequest,
		success func(schema.StatusCode, schema.Response[schema.Session]),
		failure func(schema.StatusCode, schema.Response[schema.Error]),
	)
	Refresh(
		request RefreshRequest,
		success func(schema.StatusCode, schema.Response[schema.Session]),
		failure func(schema.StatusCode, schema.Response[schema.Error]),
	)
	UpdateEmail(
		subject schema.UserId,
		request UpdateEmailRequest,
		success func(schema.StatusCode, schema.Response[schema.Session]),
		failure func(schema.StatusCode, schema.Response[schema.Error]),
	)
	UpdatePassword(
		subject schema.UserId,
		request UpdatePasswordRequest,
		success func(schema.StatusCode, schema.Response[schema.Session]),
		failure func(schema.StatusCode, schema.Response[schema.Error]),
	)
	RegisterForPushNotifications(
		subject schema.UserId,
		request RegisterForPushNotificationsRequest,
		success func(schema.StatusCode, schema.VoidResponse),
		failure func(schema.StatusCode, schema.Response[schema.Error]),
	)
	Logout(
		subject schema.UserId,
		success func(schema.StatusCode, schema.VoidResponse),
		failure func(schema.StatusCode, schema.Response[schema.Error]),
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
