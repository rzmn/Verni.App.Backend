package spendings

type GetDealsErrorCode int

const (
	_ GetDealsErrorCode = iota
	GetDealsErrorInternal
)

func (c GetDealsErrorCode) Message() string {
	switch c {
	case GetDealsErrorInternal:
		return "internal error"
	default:
		return "unknown error"
	}
}
