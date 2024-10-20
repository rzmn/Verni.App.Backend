package friends

type AcceptFriendRequestErrorCode int

const (
	_ AcceptFriendRequestErrorCode = iota
	AcceptFriendRequestErrorNoSuchRequest
	AcceptFriendRequestErrorAlreadyFriends
	AcceptFriendRequestErrorInternal
)

func (c AcceptFriendRequestErrorCode) Message() string {
	switch c {
	case AcceptFriendRequestErrorNoSuchRequest:
		return "no such request"
	case AcceptFriendRequestErrorAlreadyFriends:
		return "already friends"
	case AcceptFriendRequestErrorInternal:
		return "internal error"
	default:
		return "unknown error"
	}
}
