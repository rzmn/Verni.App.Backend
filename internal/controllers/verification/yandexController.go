package verification

import (
	"fmt"
	"log"
	"math/rand"
	"net/smtp"
	"verni/internal/common"
	"verni/internal/repositories/auth"
)

type yandexController struct {
	verification VerificationRepository
	auth         AuthRepository
	sender       string
	password     string
	host         string
	port         string
}

func (s *yandexController) SendConfirmationCode(uid UserId) *common.CodeBasedError[SendConfirmationCodeErrorCode] {
	const op = "confirmation.EmailConfirmation.SendConfirmationCode"
	log.Printf("%s: start[uid=%s]", op, uid)
	user, err := s.auth.GetUserInfo(auth.UserId(uid))
	if err != nil {
		log.Printf("%s: cannot get user email by id err: %v", op, err)
		return common.NewErrorWithDescription(SendConfirmationCodeErrorInternal, err.Error())
	}
	email := user.Email
	code := fmt.Sprintf("%d", generate6DigitCode())
	transaction := s.verification.StoreEmailVerificationCode(email, code)
	if err := transaction.Perform(); err != nil {
		log.Printf("%s: store tokens failed %v", op, err)
		return common.NewErrorWithDescription(SendConfirmationCodeErrorInternal, err.Error())
	}
	to := []string{
		email,
	}
	message := []byte(fmt.Sprintf("From: Verni <%s>\r\n", s.sender) +
		fmt.Sprintf("To: %s\r\n", email) +
		"Subject: Confirm your Verni email\r\n" +
		"\r\n" +
		fmt.Sprintf("Email Verification code: %s.\r\n", code),
	)
	auth := smtp.PlainAuth("", s.sender, s.password, s.host)
	err = smtp.SendMail(s.host+":"+s.port, auth, s.sender, to, message)
	if err != nil {
		log.Printf("%s: send failed: %v", op, err)
		transaction.Rollback()
		return common.NewErrorWithDescription(SendConfirmationCodeErrorNotDelivered, err.Error())
	}
	log.Printf("%s: success[uid=%s]", op, uid)
	return nil
}

func (s *yandexController) ConfirmEmail(uid UserId, code string) *common.CodeBasedError[ConfirmEmailErrorCode] {
	const op = "confirmation.EmailConfirmation.ConfirmEmail"
	log.Printf("%s: start[uid=%s]", op, uid)
	user, err := s.auth.GetUserInfo(auth.UserId(uid))
	if err != nil {
		log.Printf("%s: cannot get user email by id err: %v", op, err)
		return common.NewErrorWithDescription(ConfirmEmailErrorInternal, err.Error())
	}
	email := user.Email
	codeFromDb, err := s.verification.GetEmailVerificationCode(email)
	if err != nil {
		log.Printf("%s: extract token failed: %v", op, err)
		return common.NewErrorWithDescription(ConfirmEmailErrorInternal, err.Error())
	}
	if codeFromDb == nil {
		log.Printf("%s: code has not been sent", op)
		return common.NewErrorWithDescription(ConfirmEmailErrorInternal, "code has not been sent")
	}
	if *codeFromDb != code {
		log.Printf("%s: verification code is wrong", op)
		return common.NewError(ConfirmEmailErrorWrongConfirmationCode)
	}
	transaction := s.auth.MarkUserEmailValidated(auth.UserId(uid))
	if err := transaction.Perform(); err != nil {
		log.Printf("%s: failed to mark email as validated: %v", op, err)
		return common.NewErrorWithDescription(ConfirmEmailErrorInternal, err.Error())
	}
	log.Printf("%s: success[uid=%s]", op, uid)
	return nil
}

func generate6DigitCode() int {
	max := 999999
	min := 100000
	return rand.Intn(max-min) + min
}
