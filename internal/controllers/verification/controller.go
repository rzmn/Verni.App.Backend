package verification

import (
	"verni/internal/common"
	authRepository "verni/internal/repositories/auth"
	verificationRepository "verni/internal/repositories/verification"
)

type UserId string
type Controller interface {
	SendConfirmationCode(uid UserId) *common.CodeBasedError[SendConfirmationCodeErrorCode]
	ConfirmEmail(uid UserId, code string) *common.CodeBasedError[ConfirmEmailErrorCode]
}

type YandexConfig struct {
	Address  string `json:"address"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     string `json:"port"`
}

type VerificationRepository verificationRepository.Repository
type AuthRepository authRepository.Repository

func YandexController(
	config YandexConfig,
	verification VerificationRepository,
	auth AuthRepository,
) Controller {
	return &yandexController{
		verification: verification,
		auth:         auth,
		sender:       config.Address,
		password:     config.Password,
		host:         config.Host,
		port:         config.Port,
	}
}
