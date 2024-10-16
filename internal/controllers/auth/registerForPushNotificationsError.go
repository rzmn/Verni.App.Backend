package auth

type RegisterForPushNotificationsErrorCode int

const (
	_ RegisterForPushNotificationsErrorCode = iota
	RegisterForPushNotificationsErrorInternal
)

func (c RegisterForPushNotificationsErrorCode) Message() string {
	switch c {
	case RegisterForPushNotificationsErrorInternal:
		return "internal error"
	default:
		return "unknown error"
	}
}
