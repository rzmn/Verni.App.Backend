package friends

type RollbackFriendRequestErrorCode int

const (
	_ RollbackFriendRequestErrorCode = iota
	RejectFriendRequestErrorNoSuchRequest
	RejectFriendRequestErrorAlreadyFriends
	RejectFriendRequestErrorInternal
)

func (c RollbackFriendRequestErrorCode) Message() string {
	switch c {
	case RejectFriendRequestErrorNoSuchRequest:
		return "no such request"
	case RejectFriendRequestErrorAlreadyFriends:
		return "already friends"
	case RejectFriendRequestErrorInternal:
		return "internal error"
	default:
		return "unknown error"
	}
}
