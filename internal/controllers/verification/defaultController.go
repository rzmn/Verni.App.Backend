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

func (c *defaultController) SendConfirmationCode(uid UserId) *common.CodeBasedError[SendConfirmationCodeErrorCode] {
	const op = "confirmation.EmailConfirmation.SendConfirmationCode"
	c.logger.LogInfo("%s: start[uid=%s]", op, uid)
	user, err := c.auth.GetUserInfo(auth.UserId(uid))
	if err != nil {
		c.logger.LogInfo("%s: cannot get user email by id err: %v", op, err)
		return common.NewErrorWithDescription(SendConfirmationCodeErrorInternal, err.Error())
	}
	email := user.Email
	code := fmt.Sprintf("%d", generate6DigitCode())
	transaction := c.verification.StoreEmailVerificationCode(email, code)
	if err := transaction.Perform(); err != nil {
		c.logger.LogInfo("%s: store tokens failed %v", op, err)
		return common.NewErrorWithDescription(SendConfirmationCodeErrorInternal, err.Error())
	}
	if err := c.emailService.Send(
		"Subject: Confirm your Verni email\r\n"+
			"\r\n"+
			fmt.Sprintf("Email Verification code: %s.\r\n", code),
		email,
	); err != nil {
		c.logger.LogInfo("%s: send failed: %v", op, err)
		transaction.Rollback()
		return common.NewErrorWithDescription(SendConfirmationCodeErrorNotDelivered, err.Error())
	}
	c.logger.LogInfo("%s: success[uid=%s]", op, uid)
	return nil
}

func (c *defaultController) ConfirmEmail(uid UserId, code string) *common.CodeBasedError[ConfirmEmailErrorCode] {
	const op = "confirmation.EmailConfirmation.ConfirmEmail"
	c.logger.LogInfo("%s: start[uid=%s]", op, uid)
	user, err := c.auth.GetUserInfo(auth.UserId(uid))
	if err != nil {
		c.logger.LogInfo("%s: cannot get user email by id err: %v", op, err)
		return common.NewErrorWithDescription(ConfirmEmailErrorInternal, err.Error())
	}
	email := user.Email
	codeFromDb, err := c.verification.GetEmailVerificationCode(email)
	if err != nil {
		c.logger.LogInfo("%s: extract token failed: %v", op, err)
		return common.NewErrorWithDescription(ConfirmEmailErrorInternal, err.Error())
	}
	if codeFromDb == nil {
		c.logger.LogInfo("%s: code has not been sent", op)
		return common.NewErrorWithDescription(ConfirmEmailErrorCodeHasNotBeenSent, "code has not been sent")
	}
	if *codeFromDb != code {
		c.logger.LogInfo("%s: verification code is wrong", op)
		return common.NewError(ConfirmEmailErrorWrongConfirmationCode)
	}
	transaction := c.auth.MarkUserEmailValidated(auth.UserId(uid))
	if err := transaction.Perform(); err != nil {
		c.logger.LogInfo("%s: failed to mark email as validated: %v", op, err)
		return common.NewErrorWithDescription(ConfirmEmailErrorInternal, err.Error())
	}
	c.logger.LogInfo("%s: success[uid=%s]", op, uid)
	return nil
}

func generate6DigitCode() int {
	max := 999999
	min := 100000
	return rand.Intn(max-min) + min
}
