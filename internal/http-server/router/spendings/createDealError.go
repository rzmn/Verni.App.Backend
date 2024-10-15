package spendings

type CreateDealErrorCode int

const (
	_ CreateDealErrorCode = iota
	CreateDealErrorNoSuchUser
	CreateDealErrorNotAFriend
	CreateDealErrorInternal
)

func (c CreateDealErrorCode) Message() string {
	switch c {
	case CreateDealErrorNoSuchUser:
		return "no such user"
	case CreateDealErrorNotAFriend:
		return "not a friend"
	case CreateDealErrorInternal:
		return "internal error"
	default:
		return "unknown error"
	}
}
