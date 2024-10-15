package profile

type UpdateDisplayNameErrorCode int

const (
	_ UpdateDisplayNameErrorCode = iota
	UpdateDisplayNameErrorNotFound
	UpdateDisplayNameErrorWrongFormat
	UpdateDisplayNameErrorInternal
)

func (c UpdateDisplayNameErrorCode) Message() string {
	switch c {
	case UpdateDisplayNameErrorNotFound:
		return "user not found"
	case UpdateDisplayNameErrorWrongFormat:
		return "wrong format"
	case UpdateDisplayNameErrorInternal:
		return "internal error"
	default:
		return "unknown error"
	}
}
