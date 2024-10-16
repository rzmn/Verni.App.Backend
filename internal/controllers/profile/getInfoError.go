package profile

type GetInfoErrorCode int

const (
	_ GetInfoErrorCode = iota
	GetInfoErrorNotFound
	GetInfoErrorInternal
)

func (c GetInfoErrorCode) Message() string {
	switch c {
	case GetInfoErrorNotFound:
		return "user not found"
	case GetInfoErrorInternal:
		return "internal error"
	default:
		return "unknown error"
	}
}
