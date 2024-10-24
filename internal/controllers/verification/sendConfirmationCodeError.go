package verification

type SendConfirmationCodeErrorCode int

const (
	_ SendConfirmationCodeErrorCode = iota
	SendConfirmationCodeErrorNotDelivered
	SendConfirmationCodeErrorInternal
)

func (c SendConfirmationCodeErrorCode) Message() string {
	switch c {
	case SendConfirmationCodeErrorNotDelivered:
		return "not delivered"
	case SendConfirmationCodeErrorInternal:
		return "internal error"
	default:
		return "unknown error"
	}
}
