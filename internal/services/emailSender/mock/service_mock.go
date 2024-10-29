package emailSender_mock

type SendCall struct {
	Subject string
	Email   string
}

type ServiceMock struct {
	SendCalls []SendCall
	SendImpl  func(subject string, email string) error
}

func (c *ServiceMock) Send(subject string, email string) error {
	return c.SendImpl(subject, email)
}
