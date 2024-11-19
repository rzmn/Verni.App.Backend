package users_mock

import (
	"github.com/rzmn/Verni.App.Backend/internal/repositories"
	"github.com/rzmn/Verni.App.Backend/internal/repositories/users"
)

type RepositoryMock struct {
	StoreUserImpl         func(user users.User) repositories.MutationWorkItem
	GetUsersImpl          func(ids []users.UserId) ([]users.User, error)
	SearchUsersImpl       func(query string) ([]users.User, error)
	UpdateDisplayNameImpl func(name string, id users.UserId) repositories.MutationWorkItem
	UpdateAvatarIdImpl    func(avatarId *users.AvatarId, id users.UserId) repositories.MutationWorkItem
}

func (c *RepositoryMock) StoreUser(user users.User) repositories.MutationWorkItem {
	return c.StoreUserImpl(user)
}

func (c *RepositoryMock) GetUsers(ids []users.UserId) ([]users.User, error) {
	return c.GetUsersImpl(ids)
}

func (c *RepositoryMock) SearchUsers(query string) ([]users.User, error) {
	return c.SearchUsersImpl(query)
}

func (c *RepositoryMock) UpdateDisplayName(name string, id users.UserId) repositories.MutationWorkItem {
	return c.UpdateDisplayNameImpl(name, id)
}

func (c *RepositoryMock) UpdateAvatarId(avatarId *users.AvatarId, id users.UserId) repositories.MutationWorkItem {
	return c.UpdateAvatarIdImpl(avatarId, id)
}
