package defaultController

import (
	"fmt"
	"math/rand"
	"verni/internal/common"
	"verni/internal/controllers/verification"
	authRepository "verni/internal/repositories/auth"
	verificationRepository "verni/internal/repositories/verification"
	"verni/internal/services/emailSender"
	"verni/internal/services/logging"
)

type VerificationRepository verificationRepository.Repository
type AuthRepository authRepository.Repository

func New(
	verification VerificationRepository,
	auth AuthRepository,
	emailService emailSender.Service,
	logger logging.Service,
) verification.Controller {
	return &defaultController{
		verification: verification,
		auth:         auth,
		emailService: emailService,
		logger:       logger,
	}
}

type defaultController struct {
	verification VerificationRepository
	auth         AuthRepository
	emailService emailSender.Service
	logger       logging.Service
}

func (c *defaultController) SendConfirmationCode(uid verification.UserId) *common.CodeBasedError[verification.SendConfirmationCodeErrorCode] {
	const op = "confirmation.EmailConfirmation.SendConfirmationCode"
	c.logger.LogInfo("%s: start[uid=%s]", op, uid)
	user, err := c.auth.GetUserInfo(authRepository.UserId(uid))
	if err != nil {
		c.logger.LogInfo("%s: cannot get user email by id err: %v", op, err)
		return common.NewErrorWithDescription(verification.SendConfirmationCodeErrorInternal, err.Error())
	}
	email := user.Email
	code := fmt.Sprintf("%d", generate6DigitCode())
	transaction := c.verification.StoreEmailVerificationCode(email, code)
	if err := transaction.Perform(); err != nil {
		c.logger.LogInfo("%s: store tokens failed %v", op, err)
		return common.NewErrorWithDescription(verification.SendConfirmationCodeErrorInternal, err.Error())
	}
	if err := c.emailService.Send(
		"Subject: Confirm your Verni email\r\n"+
			"\r\n"+
			fmt.Sprintf("Email Verification code: %s.\r\n", code),
		email,
	); err != nil {
		c.logger.LogInfo("%s: send failed: %v", op, err)
		transaction.Rollback()
		return common.NewErrorWithDescription(verification.SendConfirmationCodeErrorNotDelivered, err.Error())
	}
	c.logger.LogInfo("%s: success[uid=%s]", op, uid)
	return nil
}

func (c *defaultController) ConfirmEmail(uid verification.UserId, code string) *common.CodeBasedError[verification.ConfirmEmailErrorCode] {
	const op = "confirmation.EmailConfirmation.ConfirmEmail"
	c.logger.LogInfo("%s: start[uid=%s]", op, uid)
	user, err := c.auth.GetUserInfo(authRepository.UserId(uid))
	if err != nil {
		c.logger.LogInfo("%s: cannot get user email by id err: %v", op, err)
		return common.NewErrorWithDescription(verification.ConfirmEmailErrorInternal, err.Error())
	}
	email := user.Email
	codeFromDb, err := c.verification.GetEmailVerificationCode(email)
	if err != nil {
		c.logger.LogInfo("%s: extract token failed: %v", op, err)
		return common.NewErrorWithDescription(verification.ConfirmEmailErrorInternal, err.Error())
	}
	if codeFromDb == nil {
		c.logger.LogInfo("%s: code has not been sent", op)
		return common.NewErrorWithDescription(verification.ConfirmEmailErrorCodeHasNotBeenSent, "code has not been sent")
	}
	if *codeFromDb != code {
		c.logger.LogInfo("%s: verification code is wrong", op)
		return common.NewError(verification.ConfirmEmailErrorWrongConfirmationCode)
	}
	transaction := c.auth.MarkUserEmailValidated(authRepository.UserId(uid))
	if err := transaction.Perform(); err != nil {
		c.logger.LogInfo("%s: failed to mark email as validated: %v", op, err)
		return common.NewErrorWithDescription(verification.ConfirmEmailErrorInternal, err.Error())
	}
	c.logger.LogInfo("%s: success[uid=%s]", op, uid)
	return nil
}

func generate6DigitCode() int {
	max := 999999
	min := 100000
	return rand.Intn(max-min) + min
}
