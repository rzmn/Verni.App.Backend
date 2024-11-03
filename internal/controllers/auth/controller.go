package auth

import (
	"verni/internal/common"
	"verni/internal/services/formatValidation"
	"verni/internal/services/jwt"
	"verni/internal/services/logging"

	authRepository "verni/internal/repositories/auth"
	pushNotificationsRepository "verni/internal/repositories/pushNotifications"
)

type UserId string
type AuthRepository authRepository.Repository
type PushTokensRepository pushNotificationsRepository.Repository

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

func DefaultController(
	authRepository AuthRepository,
	pushTokensRepository PushTokensRepository,
	jwtService jwt.Service,
	formatValidationService formatValidation.Service,
	logger logging.Service,
) Controller {
	return &defaultController{
		authRepository:          authRepository,
		pushTokensRepository:    pushTokensRepository,
		jwtService:              jwtService,
		formatValidationService: formatValidationService,
		logger:                  logger,
	}
}
