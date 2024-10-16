package auth

type UpdateEmailErrorCode int

const (
	_ UpdateEmailErrorCode = iota
	UpdateEmailErrorWrongFormat
	UpdateEmailErrorAlreadyTaken
	UpdateEmailErrorInternal
)

func (c UpdateEmailErrorCode) Message() string {
	switch c {
	case UpdateEmailErrorWrongFormat:
		return "wrong format"
	case UpdateEmailErrorAlreadyTaken:
		return "already taken"
	case UpdateEmailErrorInternal:
		return "internal error"
	default:
		return "unknown error"
	}
}
