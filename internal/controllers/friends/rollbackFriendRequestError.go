package friends

type RollbackFriendRequestErrorCode int

const (
	_ RollbackFriendRequestErrorCode = iota
	RollbackFriendRequestErrorNoSuchRequest
	RollbackFriendRequestErrorAlreadyFriends
	RollbackFriendRequestErrorInternal
)

func (c RollbackFriendRequestErrorCode) Message() string {
	switch c {
	case RollbackFriendRequestErrorNoSuchRequest:
		return "no such request"
	case RollbackFriendRequestErrorAlreadyFriends:
		return "already friends"
	case RollbackFriendRequestErrorInternal:
		return "internal error"
	default:
		return "unknown error"
	}
}
