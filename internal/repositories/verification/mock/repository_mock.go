package verification_mock

import "github.com/rzmn/governi/internal/repositories"

type RepositoryMock struct {
	StoreEmailVerificationCodeImpl  func(email string, code string) repositories.MutationWorkItem
	GetEmailVerificationCodeImpl    func(email string) (*string, error)
	RemoveEmailVerificationCodeImpl func(email string) repositories.MutationWorkItem
}

func (c *RepositoryMock) StoreEmailVerificationCode(email string, code string) repositories.MutationWorkItem {
	return c.StoreEmailVerificationCodeImpl(email, code)
}
func (c *RepositoryMock) GetEmailVerificationCode(email string) (*string, error) {
	return c.GetEmailVerificationCodeImpl(email)
}
func (c *RepositoryMock) RemoveEmailVerificationCode(email string) repositories.MutationWorkItem {
	return c.RemoveEmailVerificationCodeImpl(email)
}
