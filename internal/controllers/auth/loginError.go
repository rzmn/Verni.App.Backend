package auth

type LoginErrorCode int

const (
	_ LoginErrorCode = iota
	LoginErrorWrongCredentials
	LoginErrorInternal
)

func (c LoginErrorCode) Message() string {
	switch c {
	case LoginErrorWrongCredentials:
		return "wrong credentials"
	case LoginErrorInternal:
		return "internal error"
	default:
		return "unknown error"
	}
}
