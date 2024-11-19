package auth

import (
	"github.com/rzmn/Verni.App.Backend/internal/common"
)

type UserId string

type Session struct {
	Id           UserId
	AccessToken  string
	RefreshToken string
}

type Controller interface {
	Signup(email string, password string) (Session, *common.CodeBasedError[SignupErrorCode])
	Login(email string, password string) (Session, *common.CodeBasedError[LoginErrorCode])
	Refresh(refreshToken string) (Session, *common.CodeBasedError[RefreshErrorCode])
	Logout(id UserId) *common.CodeBasedError[LogoutErrorCode]

	UpdateEmail(email string, id UserId) (Session, *common.CodeBasedError[UpdateEmailErrorCode])
	UpdatePassword(oldPassword string, newPassword string, id UserId) (Session, *common.CodeBasedError[UpdatePasswordErrorCode])

	RegisterForPushNotifications(pushToken string, id UserId) *common.CodeBasedError[RegisterForPushNotificationsErrorCode]
}
