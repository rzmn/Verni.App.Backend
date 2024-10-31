package pushNotifications_mock

import (
	"verni/internal/repositories"
	"verni/internal/repositories/pushNotifications"
)

type RepositoryMock struct {
	StorePushTokenImpl func(uid pushNotifications.UserId, token string) repositories.MutationWorkItem
	GetPushTokenImpl   func(uid pushNotifications.UserId) (*string, error)
}

func (c *RepositoryMock) StorePushToken(uid pushNotifications.UserId, token string) repositories.MutationWorkItem {
	return c.StorePushTokenImpl(uid, token)
}

func (c *RepositoryMock) GetPushToken(uid pushNotifications.UserId) (*string, error) {
	return c.GetPushTokenImpl(uid)
}
