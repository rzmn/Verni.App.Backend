package auth

type UpdatePasswordErrorCode int

const (
	_ UpdatePasswordErrorCode = iota
	UpdatePasswordErrorOldPasswordIsWrong
	UpdatePasswordErrorWrongFormat
	UpdatePasswordErrorInternal
)

func (c UpdatePasswordErrorCode) Message() string {
	switch c {
	case UpdatePasswordErrorOldPasswordIsWrong:
		return "old password is wrong"
	case UpdatePasswordErrorInternal:
		return "internal error"
	default:
		return "unknown error"
	}
}
