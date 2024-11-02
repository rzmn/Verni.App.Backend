package spendings

type AddExpenseErrorCode int

const (
	_ AddExpenseErrorCode = iota
	AddExpenseErrorNoSuchUser
	AddExpenseErrorNotYourExpense
	AddExpenseErrorInternal
)

func (c AddExpenseErrorCode) Message() string {
	switch c {
	case AddExpenseErrorNoSuchUser:
		return "no such user"
	case AddExpenseErrorNotYourExpense:
		return "not your expense"
	case AddExpenseErrorInternal:
		return "internal error"
	default:
		return "unknown error"
	}
}
