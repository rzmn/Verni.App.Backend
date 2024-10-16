package spendings

type GetDealsErrorCode int

const (
	_ GetDealsErrorCode = iota
	GetDealsErrorNoSuchUser
	GetDealsErrorInternal
)

func (c GetDealsErrorCode) Message() string {
	switch c {
	case GetDealsErrorNoSuchUser:
		return "no such user"
	case GetDealsErrorInternal:
		return "internal error"
	default:
		return "unknown error"
	}
}
