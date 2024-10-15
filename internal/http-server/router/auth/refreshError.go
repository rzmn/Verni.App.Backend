package auth

type RefreshErrorCode int

const (
	_ RefreshErrorCode = iota
	RefreshErrorTokenExpired
	RefreshErrorTokenIsWrong
	RefreshErrorInternal
)

func (c RefreshErrorCode) Message() string {
	switch c {
	case RefreshErrorTokenExpired:
		return "token expired"
	case RefreshErrorTokenIsWrong:
		return "token is wrong"
	case RefreshErrorInternal:
		return "internal error"
	default:
		return "unknown error"
	}
}
