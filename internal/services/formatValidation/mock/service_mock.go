package formatValidation_mock

type ServiceMock struct {
	ValidateEmailFormatImpl       func(email string) error
	ValidatePasswordFormatImpl    func(password string) error
	ValidateDisplayNameFormatImpl func(name string) error
}

func (s *ServiceMock) ValidateEmailFormat(email string) error {
	return s.ValidateEmailFormatImpl(email)
}

func (s *ServiceMock) ValidatePasswordFormat(password string) error {
	return s.ValidatePasswordFormatImpl(password)
}

func (s *ServiceMock) ValidateDisplayNameFormat(name string) error {
	return s.ValidateDisplayNameFormatImpl(name)
}
