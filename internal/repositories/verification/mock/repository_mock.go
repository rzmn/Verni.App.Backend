package verification_mock

import "verni/internal/repositories"

type StoreEmailVerificationCodeCall struct {
	Email string
	Code  string
}

type RepositoryMock struct {
	StoreEmailVerificationCodeCalls  []StoreEmailVerificationCodeCall
	GetEmailVerificationCodeCalls    []string
	RemoveEmailVerificationCodeCalls []string

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
