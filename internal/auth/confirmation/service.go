package confirmation

import (
	"verni/internal/common"
	"verni/internal/storage"
)

type Service interface {
	SendConfirmationCode(email string) *common.CodeBasedError[SendConfirmationCodeErrorCode]
	ConfirmEmail(email string, code string) *common.CodeBasedError[ConfirmEmailErrorCode]
}

func YandexService(storage storage.Storage, sender string, password string, host string, port string) Service {
	return &defaultService{
		Storage:  storage,
		Sender:   sender,
		Password: password,
		Host:     host,
		Port:     port,
	}
}
