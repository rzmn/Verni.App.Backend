package friends

type AcceptFriendRequestErrorCode int

const (
	_ AcceptFriendRequestErrorCode = iota
	AcceptFriendRequestErrorNoSuchRequest
	AcceptFriendRequestErrorInternal
)

func (c AcceptFriendRequestErrorCode) Message() string {
	switch c {
	case AcceptFriendRequestErrorNoSuchRequest:
		return "no such request"
	case AcceptFriendRequestErrorInternal:
		return "internal error"
	default:
		return "unknown error"
	}
}
