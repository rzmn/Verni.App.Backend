package friends

type RejectFriendRequestErrorCode int

const (
	_ RejectFriendRequestErrorCode = iota
	RejectFriendRequestErrorNoSuchRequest
	RejectFriendRequestErrorInternal
)

func (c RejectFriendRequestErrorCode) Message() string {
	switch c {
	case RejectFriendRequestErrorNoSuchRequest:
		return "no such request"
	case RejectFriendRequestErrorInternal:
		return "internal error"
	default:
		return "unknown error"
	}
}
