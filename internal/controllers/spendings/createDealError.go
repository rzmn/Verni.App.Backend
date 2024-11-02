package spendings

type CreateDealErrorCode int

const (
	_ CreateDealErrorCode = iota
	CreateDealErrorNoSuchUser
	CreateDealErrorNotYourExpense
	CreateDealErrorInternal
)

func (c CreateDealErrorCode) Message() string {
	switch c {
	case CreateDealErrorNoSuchUser:
		return "no such user"
	case CreateDealErrorNotYourExpense:
		return "not your expense"
	case CreateDealErrorInternal:
		return "internal error"
	default:
		return "unknown error"
	}
}
