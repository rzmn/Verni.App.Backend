package verification

import (
	"verni/internal/common"
	authRepository "verni/internal/repositories/auth"
	verificationRepository "verni/internal/repositories/verification"
	"verni/internal/services/emailSender"
)

type UserId string
type Controller interface {
	SendConfirmationCode(uid UserId) *common.CodeBasedError[SendConfirmationCodeErrorCode]
	ConfirmEmail(uid UserId, code string) *common.CodeBasedError[ConfirmEmailErrorCode]
}

type VerificationRepository verificationRepository.Repository
type AuthRepository authRepository.Repository

func YandexController(
	verification VerificationRepository,
	auth AuthRepository,
	emailService emailSender.Service,
) Controller {
	return &yandexController{
		verification: verification,
		auth:         auth,
		emailService: emailService,
	}
}
