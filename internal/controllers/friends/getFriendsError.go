package friends

type GetFriendsErrorCode int

const (
	_ GetFriendsErrorCode = iota
	GetFriendsErrorInternal
)

func (c GetFriendsErrorCode) Message() string {
	switch c {
	case GetFriendsErrorInternal:
		return "internal error"
	default:
		return "unknown error"
	}
}
