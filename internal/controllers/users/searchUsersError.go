package users

type SearchUsersErrorCode int

const (
	_ SearchUsersErrorCode = iota
	SearchUsersErrorInternal
)

func (c SearchUsersErrorCode) Message() string {
	switch c {
	case SearchUsersErrorInternal:
		return "internal error"
	default:
		return "unknown error"
	}
}
