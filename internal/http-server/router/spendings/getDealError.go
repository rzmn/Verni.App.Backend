package spendings

type GetDealErrorCode int

const (
	_ GetDealErrorCode = iota
	GetDealErrorDealNotFound
	GetDealErrorNotAFriend
	GetDealErrorNotYourDeal
	GetDealErrorInternal
)

func (c GetDealErrorCode) Message() string {
	switch c {
	case GetDealErrorDealNotFound:
		return "deal not found"
	case GetDealErrorNotYourDeal:
		return "not your deal"
	case GetDealErrorInternal:
		return "internal error"
	default:
		return "unknown error"
	}
}
