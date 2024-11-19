package verification

import (
	"github.com/rzmn/governi/internal/repositories"
)

type Repository interface {
	StoreEmailVerificationCode(email string, code string) repositories.MutationWorkItem
	GetEmailVerificationCode(email string) (*string, error)
	RemoveEmailVerificationCode(email string) repositories.MutationWorkItem
}
