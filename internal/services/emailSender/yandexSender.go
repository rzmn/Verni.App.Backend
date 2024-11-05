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

func (c *yandexService) Send(subject string, email string) error {
	const op = "emailSender.yandexService.Send"
	c.logger.LogInfo("%s: start", op)
	to := []string{
		email,
	}
	auth := smtp.PlainAuth("", c.sender, c.password, c.host)

	message := []byte(
		fmt.Sprintf("From: Verni <%s>\r\n", c.sender) +
			fmt.Sprintf("To: %s\r\n", email) + subject,
	)
	err := smtp.SendMail(c.host+":"+c.port, auth, c.sender, to, []byte(message))
	if err != nil {
		c.logger.LogInfo("%s: send failed: %v", op, err)
		return err
	}
	c.logger.LogInfo("%s: success", op)
	return nil
}
