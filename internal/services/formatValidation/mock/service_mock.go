package formatValidation_mock

type ServiceMock struct {
	ValidateEmailFormatCalls    []string
	ValidatePasswordFormatCalls []string

	ValidateEmailFormatImpl    func(email string) error
	ValidatePasswordFormatImpl func(password string) error
}

func (s *ServiceMock) ValidateEmailFormat(email string) error {
	return s.ValidateEmailFormatImpl(email)
}

func (s *ServiceMock) ValidatePasswordFormat(password string) error {
	return s.ValidatePasswordFormatImpl(password)
}
