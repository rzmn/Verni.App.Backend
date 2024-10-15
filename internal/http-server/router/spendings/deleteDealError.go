package spendings

type DeleteDealErrorCode int

const (
	_ DeleteDealErrorCode = iota
	DeleteDealErrorDealNotFound
	DeleteDealErrorNotAFriend
	DeleteDealErrorNotYourDeal
	DeleteDealErrorInternal
)

func (c DeleteDealErrorCode) Message() string {
	switch c {
	case DeleteDealErrorDealNotFound:
		return "deal not found"
	case DeleteDealErrorNotAFriend:
		return "not a friend"
	case DeleteDealErrorNotYourDeal:
		return "not your deal"
	case DeleteDealErrorInternal:
		return "internal error"
	default:
		return "unknown error"
	}
}
