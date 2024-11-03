package emailSender

import (
	"fmt"
	"net/smtp"
	"verni/internal/services/logging"
)

type yandexService struct {
	sender   string
	password string
	host     string
	port     string
	logger   logging.Service
}

func (s *yandexService) Send(subject string, email string) error {
	const op = "emailSender.yandexService.Send"
	s.logger.Log("%s: start", op)
	to := []string{
		email,
	}
	auth := smtp.PlainAuth("", s.sender, s.password, s.host)

	message := []byte(
		fmt.Sprintf("From: Verni <%s>\r\n", s.sender) +
			fmt.Sprintf("To: %s\r\n", email) + subject,
	)
	err := smtp.SendMail(s.host+":"+s.port, auth, s.sender, to, []byte(message))
	if err != nil {
		s.logger.Log("%s: send failed: %v", op, err)
		return err
	}
	s.logger.Log("%s: success", op)
	return nil
}
