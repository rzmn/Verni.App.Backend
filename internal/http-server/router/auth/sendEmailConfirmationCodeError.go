package auth

type SendEmailConfirmationCodeErrorCode int

const (
	_ SendEmailConfirmationCodeErrorCode = iota
	SendEmailConfirmationCodeErrorAlreadyConfirmed
	SendEmailConfirmationCodeErrorNotDelivered
	SendEmailConfirmationCodeErrorInternal
)

func (c SendEmailConfirmationCodeErrorCode) Message() string {
	switch c {
	case SendEmailConfirmationCodeErrorInternal:
		return "internal error"
	default:
		return "unknown error"
	}
}
