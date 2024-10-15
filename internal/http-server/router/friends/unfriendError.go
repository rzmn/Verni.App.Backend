package friends

type UnfriendErrorCode int

const (
	_ UnfriendErrorCode = iota
	UnfriendErrorNotAFriend
	UnfriendErrorNoSuchUser
	UnfriendErrorInternal
)

func (c UnfriendErrorCode) Message() string {
	switch c {
	case UnfriendErrorNotAFriend:
		return "not a friend"
	case UnfriendErrorNoSuchUser:
		return "no such user"
	case UnfriendErrorInternal:
		return "internal error"
	default:
		return "unknown error"
	}
}
