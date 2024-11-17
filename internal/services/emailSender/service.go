package emailSender

type Service interface {
	Send(subject string, email string) error
}
