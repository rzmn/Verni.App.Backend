package auth

type LoginErrorCode int

const (
	_ LoginErrorCode = iota
	LoginErrorInternal
	LoginErrorWrongCredentials
)

func (c LoginErrorCode) Message() string {
	switch c {
	case LoginErrorInternal:
		return "internal error"
	case LoginErrorWrongCredentials:
		return "wrong credentials"
	default:
		return "unknown error"
	}
}
