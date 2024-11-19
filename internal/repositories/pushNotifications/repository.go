package pushNotifications

import (
	"github.com/rzmn/Verni.App.Backend/internal/repositories"
)

type UserId string

type Repository interface {
	StorePushToken(uid UserId, token string) repositories.MutationWorkItem
	GetPushToken(uid UserId) (*string, error)
}
