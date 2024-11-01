package emailSender_mock

type ServiceMock struct {
	SendImpl func(subject string, email string) error
}

func (c *ServiceMock) Send(subject string, email string) error {
	return c.SendImpl(subject, email)
}
