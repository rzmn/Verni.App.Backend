package profile

type UpdateAvatarErrorCode int

const (
	_ UpdateAvatarErrorCode = iota
	UpdateAvatarErrorInternal
)

func (c UpdateAvatarErrorCode) Message() string {
	switch c {
	case UpdateAvatarErrorInternal:
		return "internal error"
	default:
		return "unknown error"
	}
}
