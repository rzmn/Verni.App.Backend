package auth

type SignupErrorCode int

const (
	_ SignupErrorCode = iota
	SignupErrorWrongFormat
	SignupErrorAlreadyTaken
	SignupErrorInternal
)

func (c SignupErrorCode) Message() string {
	switch c {
	case SignupErrorInternal:
		return "internal error"
	case SignupErrorAlreadyTaken:
		return "already taken"
	default:
		return "unknown error"
	}
}
