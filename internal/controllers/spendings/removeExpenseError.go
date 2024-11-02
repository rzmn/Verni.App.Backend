package spendings

type RemoveExpenseErrorCode int

const (
	_ RemoveExpenseErrorCode = iota
	RemoveExpenseErrorExpenseNotFound
	RemoveExpenseErrorNotAFriend
	RemoveExpenseErrorNotYourExpense
	RemoveExpenseErrorInternal
)

func (c RemoveExpenseErrorCode) Message() string {
	switch c {
	case RemoveExpenseErrorExpenseNotFound:
		return "expense not found"
	case RemoveExpenseErrorNotAFriend:
		return "not a friend"
	case RemoveExpenseErrorNotYourExpense:
		return "not your expense"
	case RemoveExpenseErrorInternal:
		return "internal error"
	default:
		return "unknown error"
	}
}
