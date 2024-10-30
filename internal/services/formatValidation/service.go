package formatValidation

type Service interface {
	ValidateEmailFormat(email string) error
	ValidatePasswordFormat(password string) error
}

func DefaultService() Service {
	return &defaultService{}
}
