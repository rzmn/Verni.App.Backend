package spendings

type GetBalanceErrorCode int

const (
	_ GetBalanceErrorCode = iota
	GetBalanceErrorInternal
)

func (c GetBalanceErrorCode) Message() string {
	switch c {
	case GetBalanceErrorInternal:
		return "internal error"
	default:
		return "unknown error"
	}
}
