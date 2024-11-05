package formatValidation_mock

type ServiceMock struct {
	ValidateEmailFormatImpl       func(email string) error
	ValidatePasswordFormatImpl    func(password string) error
	ValidateDisplayNameFormatImpl func(name string) error
}

func (c *ServiceMock) ValidateEmailFormat(email string) error {
	return c.ValidateEmailFormatImpl(email)
}

func (c *ServiceMock) ValidatePasswordFormat(password string) error {
	return c.ValidatePasswordFormatImpl(password)
}

func (c *ServiceMock) ValidateDisplayNameFormat(name string) error {
	return c.ValidateDisplayNameFormatImpl(name)
}
