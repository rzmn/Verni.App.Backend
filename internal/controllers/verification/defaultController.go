package verification

import (
	"fmt"
	"math/rand"
	"verni/internal/common"
	"verni/internal/repositories/auth"
	"verni/internal/services/emailSender"
	"verni/internal/services/logging"
)

type defaultController struct {
	verification VerificationRepository
	auth         AuthRepository
	emailService emailSender.Service
	logger       logging.Service
}

func (s *defaultController) SendConfirmationCode(uid UserId) *common.CodeBasedError[SendConfirmationCodeErrorCode] {
	const op = "confirmation.EmailConfirmation.SendConfirmationCode"
	s.logger.Log("%s: start[uid=%s]", op, uid)
	user, err := s.auth.GetUserInfo(auth.UserId(uid))
	if err != nil {
		s.logger.Log("%s: cannot get user email by id err: %v", op, err)
		return common.NewErrorWithDescription(SendConfirmationCodeErrorInternal, err.Error())
	}
	email := user.Email
	code := fmt.Sprintf("%d", generate6DigitCode())
	transaction := s.verification.StoreEmailVerificationCode(email, code)
	if err := transaction.Perform(); err != nil {
		s.logger.Log("%s: store tokens failed %v", op, err)
		return common.NewErrorWithDescription(SendConfirmationCodeErrorInternal, err.Error())
	}
	if err := s.emailService.Send(
		"Subject: Confirm your Verni email\r\n"+
			"\r\n"+
			fmt.Sprintf("Email Verification code: %s.\r\n", code),
		email,
	); err != nil {
		s.logger.Log("%s: send failed: %v", op, err)
		transaction.Rollback()
		return common.NewErrorWithDescription(SendConfirmationCodeErrorNotDelivered, err.Error())
	}
	s.logger.Log("%s: success[uid=%s]", op, uid)
	return nil
}

func (s *defaultController) ConfirmEmail(uid UserId, code string) *common.CodeBasedError[ConfirmEmailErrorCode] {
	const op = "confirmation.EmailConfirmation.ConfirmEmail"
	s.logger.Log("%s: start[uid=%s]", op, uid)
	user, err := s.auth.GetUserInfo(auth.UserId(uid))
	if err != nil {
		s.logger.Log("%s: cannot get user email by id err: %v", op, err)
		return common.NewErrorWithDescription(ConfirmEmailErrorInternal, err.Error())
	}
	email := user.Email
	codeFromDb, err := s.verification.GetEmailVerificationCode(email)
	if err != nil {
		s.logger.Log("%s: extract token failed: %v", op, err)
		return common.NewErrorWithDescription(ConfirmEmailErrorInternal, err.Error())
	}
	if codeFromDb == nil {
		s.logger.Log("%s: code has not been sent", op)
		return common.NewErrorWithDescription(ConfirmEmailErrorCodeHasNotBeenSent, "code has not been sent")
	}
	if *codeFromDb != code {
		s.logger.Log("%s: verification code is wrong", op)
		return common.NewError(ConfirmEmailErrorWrongConfirmationCode)
	}
	transaction := s.auth.MarkUserEmailValidated(auth.UserId(uid))
	if err := transaction.Perform(); err != nil {
		s.logger.Log("%s: failed to mark email as validated: %v", op, err)
		return common.NewErrorWithDescription(ConfirmEmailErrorInternal, err.Error())
	}
	s.logger.Log("%s: success[uid=%s]", op, uid)
	return nil
}

func generate6DigitCode() int {
	max := 999999
	min := 100000
	return rand.Intn(max-min) + min
}
