package spendings

type DeleteDealErrorCode int

const (
	_ DeleteDealErrorCode = iota
	DeleteDealErrorDealNotFound
	DeleteDealErrorNotAFriend
	DeleteDealErrorNotYourExpense
	DeleteDealErrorInternal
)

func (c DeleteDealErrorCode) Message() string {
	switch c {
	case DeleteDealErrorDealNotFound:
		return "deal not found"
	case DeleteDealErrorNotAFriend:
		return "not a friend"
	case DeleteDealErrorNotYourExpense:
		return "not your expense"
	case DeleteDealErrorInternal:
		return "internal error"
	default:
		return "unknown error"
	}
}
