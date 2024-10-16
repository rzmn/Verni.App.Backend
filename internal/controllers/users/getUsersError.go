package users

type GetUsersErrorCode int

const (
	_ GetUsersErrorCode = iota
	GetUsersErrorInternal
)

func (c GetUsersErrorCode) Message() string {
	switch c {
	case GetUsersErrorInternal:
		return "internal error"
	default:
		return "unknown error"
	}
}
