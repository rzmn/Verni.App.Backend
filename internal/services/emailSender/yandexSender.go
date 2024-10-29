package emailSender

import (
	"fmt"
	"log"
	"net/smtp"
)

type yandexService struct {
	sender   string
	password string
	host     string
	port     string
}

func (s *yandexService) Send(subject string, email string) error {
	const op = "emailSender.yandexService.Send"
	log.Printf("%s: start", op)
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
		log.Printf("%s: send failed: %v", op, err)
		return err
	}
	log.Printf("%s: success", op)
	return nil
}
