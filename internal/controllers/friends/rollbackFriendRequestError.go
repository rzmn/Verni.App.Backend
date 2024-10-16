package friends

type RollbackFriendRequestErrorCode int

const (
	_ RollbackFriendRequestErrorCode = iota
	RollbackFriendRequestErrorNoSuchRequest
	RollbackFriendRequestErrorInternal
)

func (c RollbackFriendRequestErrorCode) Message() string {
	switch c {
	case RollbackFriendRequestErrorNoSuchRequest:
		return "no such request"
	case RollbackFriendRequestErrorInternal:
		return "internal error"
	default:
		return "unknown error"
	}
}
