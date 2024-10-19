package auth

import (
	"verni/internal/auth/confirmation"
	"verni/internal/auth/jwt"
	"verni/internal/common"
	"verni/internal/storage"
)

type UserId storage.UserId
type Session storage.AuthenticatedSession

type Controller interface {
	Signup(email string, password string) (Session, *common.CodeBasedError[SignupErrorCode])
	Login(email string, password string) (Session, *common.CodeBasedError[LoginErrorCode])
	Refresh(refreshToken string) (Session, *common.CodeBasedError[RefreshErrorCode])
	Logout(id UserId) *common.CodeBasedError[LogoutErrorCode]

	UpdateEmail(email string, id UserId) (Session, *common.CodeBasedError[UpdateEmailErrorCode])
	UpdatePassword(oldPassword string, newPassword string, id UserId) (Session, *common.CodeBasedError[UpdatePasswordErrorCode])

	SendEmailConfirmationCode(id UserId) *common.CodeBasedError[SendEmailConfirmationCodeErrorCode]
	ConfirmEmail(code string, id UserId) *common.CodeBasedError[ConfirmEmailErrorCode]

	RegisterForPushNotifications(pushToken string, id UserId) *common.CodeBasedError[RegisterForPushNotificationsErrorCode]
}

func DefaultController(
	storage storage.Storage,
	jwtService jwt.Service,
	confirmation confirmation.Service,
) Controller {
	return &defaultController{
		storage:      storage,
		jwtService:   jwtService,
		confirmation: confirmation,
	}
}
