package friends

type UnfriendErrorCode int

const (
	_ UnfriendErrorCode = iota
	UnfriendErrorNotAFriend
	UnfriendErrorInternal
)

func (c UnfriendErrorCode) Message() string {
	switch c {
	case UnfriendErrorNotAFriend:
		return "not a friend"
	case UnfriendErrorInternal:
		return "internal error"
	default:
		return "unknown error"
	}
}
