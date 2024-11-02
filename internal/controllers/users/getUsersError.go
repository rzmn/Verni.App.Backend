package users

type GetUsersErrorCode int

const (
	_ GetUsersErrorCode = iota
	GetUsersUserNotFound
	GetUsersErrorInternal
)

func (c GetUsersErrorCode) Message() string {
	switch c {
	case GetUsersErrorInternal:
		return "internal error"
	case GetUsersUserNotFound:
		return "user not found"
	default:
		return "unknown error"
	}
}
