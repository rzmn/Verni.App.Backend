package confirmation

import (
	"accounty/internal/storage"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/smtp"
	"os"
)

type EmailConfirmation struct {
	Storage storage.Storage
}

var (
	ErrNotDeliveded    = errors.New("code is not delivered")
	ErrCodeDidNotMatch = errors.New("code did not match")
	ErrInternal        = errors.New("internal error")
)

func (e *EmailConfirmation) SendConfirmationCode(email string) error {
	const op = "confirmation.EmailConfirmation.SendConfirmationCode"

	code := fmt.Sprintf("%d", generate6DigitCode())
	if err := e.Storage.StoreEmailValidationToken(email, code); err != nil {
		log.Printf("%s: store tokens failed %v", op, err)
		return ErrInternal
	}

	from := os.Getenv("EMAIL_CODE_SENDER")
	password := os.Getenv("EMAIL_CODE_SENDER_PWD")
	to := []string{
		email,
	}
	smtpHost := os.Getenv("EMAIL_CODE_SENDER_HOST")
	smtpPort := os.Getenv("EMAIL_CODE_SENDER_PORT")

	message := []byte(fmt.Sprintf("From: Splitdumb <%s>\r\n", from) +
		fmt.Sprintf("To: %s\r\n", email) +
		"Subject: Confirm your Splitdumb email\r\n" +
		"\r\n" +
		fmt.Sprintf("Email Verification code: %s.\r\n", code),
	)
	auth := smtp.PlainAuth("", from, password, smtpHost)
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, message)
	if err != nil {
		log.Printf("%s: send failed: %v", op, err)
		_, _ = e.Storage.ExtractEmailValidationToken(email)
		return ErrNotDeliveded
	}
	fmt.Println("Email Sent Successfully!")
	return nil
}

func (e *EmailConfirmation) ConfirmEmail(email string, code string) error {
	const op = "confirmation.EmailConfirmation.ConfirmEmail"

	sentCode, err := e.Storage.ExtractEmailValidationToken(email)
	if err != nil {
		log.Printf("%s: extract token failed: %v", op, err)
		return ErrInternal
	}
	if sentCode == nil {
		log.Printf("%s: code has not been sent: %v", op, err)
		return ErrInternal
	}
	if *sentCode != code {
		return ErrCodeDidNotMatch
	}
	if err := e.Storage.ValidateEmail(email); err != nil {
		log.Printf("%s: failed to mark email as validated: %v", op, err)
		return ErrInternal
	}
	return nil
}

func generate6DigitCode() int {
	max := 999999
	min := 100000
	return rand.Intn(max-min) + min
}
