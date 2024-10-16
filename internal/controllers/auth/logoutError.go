package auth

type LogoutErrorCode int

const (
	_ LogoutErrorCode = iota
	LogoutErrorInternal
)

func (c LogoutErrorCode) Message() string {
	switch c {
	case LogoutErrorInternal:
		return "internal error"
	default:
		return "unknown error"
	}
}
