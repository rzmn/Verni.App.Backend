package verification

import (
	"github.com/rzmn/governi/internal/common"
)

type UserId string
type Controller interface {
	SendConfirmationCode(uid UserId) *common.CodeBasedError[SendConfirmationCodeErrorCode]
	ConfirmEmail(uid UserId, code string) *common.CodeBasedError[ConfirmEmailErrorCode]
}
