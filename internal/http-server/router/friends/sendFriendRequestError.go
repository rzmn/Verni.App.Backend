package friends

type SendFriendRequestErrorCode int

const (
	_ SendFriendRequestErrorCode = iota
	SendFriendRequestErrorAlreadySent
	SendFriendRequestErrorHaveIncomingRequest
	SendFriendRequestErrorAlreadyFriends
	SendFriendRequestErrorNoSuchUser
	SendFriendRequestErrorInternal
)

func (c SendFriendRequestErrorCode) Message() string {
	switch c {
	case SendFriendRequestErrorAlreadySent:
		return "friend request already sent"
	case SendFriendRequestErrorHaveIncomingRequest:
		return "already have incoming friend request"
	case SendFriendRequestErrorAlreadyFriends:
		return "already friends"
	case SendFriendRequestErrorNoSuchUser:
		return "no such user"
	case SendFriendRequestErrorInternal:
		return "internal error"
	default:
		return "unknown error"
	}
}
