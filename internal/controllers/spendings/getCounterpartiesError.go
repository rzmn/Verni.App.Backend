package spendings

type GetCounterpartiesErrorCode int

const (
	_ GetCounterpartiesErrorCode = iota
	GetCounterpartiesErrorInternal
)

func (c GetCounterpartiesErrorCode) Message() string {
	switch c {
	case GetCounterpartiesErrorInternal:
		return "internal error"
	default:
		return "unknown error"
	}
}
