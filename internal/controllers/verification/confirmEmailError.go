package verification

type ConfirmEmailErrorCode int

const (
	_ ConfirmEmailErrorCode = iota
	ConfirmEmailErrorWrongConfirmationCode
	ConfirmEmailErrorInternal
)

func (c ConfirmEmailErrorCode) Message() string {
	switch c {
	case ConfirmEmailErrorWrongConfirmationCode:
		return "wrong confirmation code"
	case ConfirmEmailErrorInternal:
		return "internal error"
	default:
		return "unknown error"
	}
}
