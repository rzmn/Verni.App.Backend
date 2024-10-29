package repository_mock

import (
	"verni/internal/repositories"

	"verni/internal/repositories/auth"
)

type CreateUserCall struct {
	Uid          auth.UserId
	Email        string
	Password     string
	RefreshToken string
}
type CheckCredentialsCall struct {
	Email    string
	Password string
}

type UpdateRefreshTokenCall struct {
	Uid   auth.UserId
	Token string
}

type UpdatePasswordCall struct {
	Uid         auth.UserId
	NewPassword string
}

type UpdateEmailCall struct {
	Uid   auth.UserId
	Email string
}

type RepositoryMock struct {
	CreateUserCalls             []CreateUserCall
	MarkUserEmailValidatedCalls []auth.UserId
	IsUserExistsCalls           []auth.UserId
	CheckCredentialsCalls       []CheckCredentialsCall
	GetUserIdByEmailCalls       []string
	UpdateRefreshTokenCalls     []UpdateRefreshTokenCall
	UpdatePasswordCalls         []UpdatePasswordCall
	UpdateEmailCalls            []UpdateEmailCall
	GetUserInfoCalls            []auth.UserId

	CreateUserImpl             func(uid auth.UserId, email string, password string, refreshToken string) repositories.MutationWorkItem
	MarkUserEmailValidatedImpl func(uid auth.UserId) repositories.MutationWorkItem
	IsUserExistsImpl           func(uid auth.UserId) (bool, error)
	CheckCredentialsImpl       func(email string, password string) (bool, error)
	GetUserIdByEmailImpl       func(email string) (*auth.UserId, error)
	UpdateRefreshTokenImpl     func(uid auth.UserId, token string) repositories.MutationWorkItem
	UpdatePasswordImpl         func(uid auth.UserId, newPassword string) repositories.MutationWorkItem
	UpdateEmailImpl            func(uid auth.UserId, newEmail string) repositories.MutationWorkItem
	GetUserInfoImpl            func(uid auth.UserId) (auth.UserInfo, error)
}

func (c *RepositoryMock) CreateUser(uid auth.UserId, email string, password string, refreshToken string) repositories.MutationWorkItem {
	return c.CreateUserImpl(uid, email, password, refreshToken)
}
func (c *RepositoryMock) MarkUserEmailValidated(uid auth.UserId) repositories.MutationWorkItem {
	return c.MarkUserEmailValidatedImpl(uid)
}
func (c *RepositoryMock) IsUserExists(uid auth.UserId) (bool, error) {
	return c.IsUserExistsImpl(uid)
}
func (c *RepositoryMock) CheckCredentials(email string, password string) (bool, error) {
	return c.CheckCredentialsImpl(email, password)
}
func (c *RepositoryMock) GetUserIdByEmail(email string) (*auth.UserId, error) {
	return c.GetUserIdByEmailImpl(email)
}
func (c *RepositoryMock) UpdateRefreshToken(uid auth.UserId, token string) repositories.MutationWorkItem {
	return c.UpdateRefreshTokenImpl(uid, token)
}
func (c *RepositoryMock) UpdatePassword(uid auth.UserId, newPassword string) repositories.MutationWorkItem {
	return c.UpdatePasswordImpl(uid, newPassword)
}
func (c *RepositoryMock) UpdateEmail(uid auth.UserId, newEmail string) repositories.MutationWorkItem {
	return c.UpdateEmailImpl(uid, newEmail)
}
func (c *RepositoryMock) GetUserInfo(uid auth.UserId) (auth.UserInfo, error) {
	return c.GetUserInfoImpl(uid)
}
