package avatars

type GetAvatarsErrorCode int

const (
	_ GetAvatarsErrorCode = iota
	GetAvatarsErrorInternal
)

func (c GetAvatarsErrorCode) Message() string {
	switch c {
	case GetAvatarsErrorInternal:
		return "internal error"
	default:
		return "unknown error"
	}
}
