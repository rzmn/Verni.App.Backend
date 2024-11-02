package spendings

type GetExpenseErrorCode int

const (
	_ GetExpenseErrorCode = iota
	GetExpenseErrorExpenseNotFound
	GetExpenseErrorNotAFriend
	GetExpenseErrorNotYourExpense
	GetExpenseErrorInternal
)

func (c GetExpenseErrorCode) Message() string {
	switch c {
	case GetExpenseErrorExpenseNotFound:
		return "expense not found"
	case GetExpenseErrorNotYourExpense:
		return "not your expense"
	case GetExpenseErrorInternal:
		return "internal error"
	default:
		return "unknown error"
	}
}
