package confirmation

import (
	"fmt"
	"log"
	"math/rand"
	"net/smtp"
	"verni/internal/common"
	"verni/internal/storage"
)

type defaultService struct {
	Storage  storage.Storage
	Sender   string
	Password string
	Host     string
	Port     string
}

func (s *defaultService) SendConfirmationCode(email string) *common.CodeBasedError[SendConfirmationCodeErrorCode] {
	const op = "confirmation.EmailConfirmation.SendConfirmationCode"
	log.Printf("%s: start", op)
	code := fmt.Sprintf("%d", generate6DigitCode())
	if err := s.Storage.StoreEmailValidationToken(email, code); err != nil {
		log.Printf("%s: store tokens failed %v", op, err)
		return common.NewErrorWithDescription(SendConfirmationCodeErrorInternal, err.Error())
	}
	to := []string{
		email,
	}
	message := []byte(fmt.Sprintf("From: Splitdumb <%s>\r\n", s.Sender) +
		fmt.Sprintf("To: %s\r\n", email) +
		"Subject: Confirm your Splitdumb email\r\n" +
		"\r\n" +
		fmt.Sprintf("Email Verification code: %s.\r\n", code),
	)
	auth := smtp.PlainAuth("", s.Sender, s.Password, s.Host)
	err := smtp.SendMail(s.Host+":"+s.Port, auth, s.Sender, to, message)
	if err != nil {
		log.Printf("%s: send failed: %v", op, err)
		_, _ = s.Storage.ExtractEmailValidationToken(email)
		return common.NewErrorWithDescription(SendConfirmationCodeErrorNotDelivered, err.Error())
	}
	log.Printf("%s: success", op)
	return nil
}

func (s *defaultService) ConfirmEmail(email string, code string) *common.CodeBasedError[ConfirmEmailErrorCode] {
	const op = "confirmation.EmailConfirmation.ConfirmEmail"
	log.Printf("%s: start", op)
	sentCode, err := s.Storage.ExtractEmailValidationToken(email)
	if err != nil {
		log.Printf("%s: extract token failed: %v", op, err)
		return common.NewErrorWithDescription(ConfirmEmailErrorInternal, err.Error())
	}
	if sentCode == nil {
		log.Printf("%s: code has not been sent", op)
		return common.NewErrorWithDescription(ConfirmEmailErrorInternal, "code has not been sent")
	}
	if *sentCode != code {
		return common.NewError(ConfirmEmailErrorWrongConfirmationCode)
	}
	if err := s.Storage.ValidateEmail(email); err != nil {
		log.Printf("%s: failed to mark email as validated: %v", op, err)
		return common.NewErrorWithDescription(ConfirmEmailErrorInternal, err.Error())
	}
	log.Printf("%s: success", op)
	return nil
}

func generate6DigitCode() int {
	max := 999999
	min := 100000
	return rand.Intn(max-min) + min
}
