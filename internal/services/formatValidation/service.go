package formatValidation

type Service interface {
	ValidateEmailFormat(email string) error
	ValidatePasswordFormat(password string) error
	ValidateDisplayNameFormat(name string) error
}

func DefaultService() Service {
	return &defaultService{}
}
