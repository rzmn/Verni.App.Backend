package formatValidation

import "verni/internal/services/logging"

type Service interface {
	ValidateEmailFormat(email string) error
	ValidatePasswordFormat(password string) error
	ValidateDisplayNameFormat(name string) error
}

func DefaultService(logger logging.Service) Service {
	return &defaultService{
		logger: logger,
	}
}
