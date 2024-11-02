package spendings

type GetExpensesErrorCode int

const (
	_ GetExpensesErrorCode = iota
	GetExpensesErrorInternal
)

func (c GetExpensesErrorCode) Message() string {
	switch c {
	case GetExpensesErrorInternal:
		return "internal error"
	default:
		return "unknown error"
	}
}
