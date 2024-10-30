package auth

import (
	"log"
	"verni/internal/common"
	"verni/internal/repositories/auth"
	"verni/internal/repositories/pushNotifications"
	"verni/internal/services/formatValidation"
	"verni/internal/services/jwt"

	"github.com/google/uuid"
)

type defaultController struct {
	authRepository          AuthRepository
	pushTokensRepository    PushTokensRepository
	jwtService              jwt.Service
	formatValidationService formatValidation.Service
}

func (c *defaultController) Signup(email string, password string) (Session, *common.CodeBasedError[SignupErrorCode]) {
	const op = "auth.defaultController.Signup"
	log.Printf("%s: start", op)
	if err := c.formatValidationService.ValidateEmailFormat(email); err != nil {
		log.Printf("%s: wrong email format err: %v", op, err)
		return Session{}, common.NewErrorWithDescription(SignupErrorWrongFormat, err.Error())
	}
	if err := c.formatValidationService.ValidatePasswordFormat(password); err != nil {
		log.Printf("%s: wrong password format err: %v", op, err)
		return Session{}, common.NewErrorWithDescription(SignupErrorWrongFormat, err.Error())
	}
	uidAccosiatedWithEmail, err := c.authRepository.GetUserIdByEmail(email)
	if err != nil {
		log.Printf("%s: getting uid by credentials from db failed err: %v", op, err)
		return Session{}, common.NewErrorWithDescription(SignupErrorInternal, err.Error())
	}
	if uidAccosiatedWithEmail != nil {
		log.Printf("%s: already has an uid accosiated with credentials", op)
		return Session{}, common.NewError(SignupErrorAlreadyTaken)
	}
	uid := uuid.New().String()
	accessToken, jwtErr := c.jwtService.IssueAccessToken(jwt.Subject(uid))
	if jwtErr != nil {
		log.Printf("%s: issuing access token failed err: %v", op, jwtErr)
		return Session{}, common.NewErrorWithDescription(SignupErrorInternal, jwtErr.Error())
	}
	refreshToken, jwtErr := c.jwtService.IssueRefreshToken(jwt.Subject(uid))
	if jwtErr != nil {
		log.Printf("%s: issuing refresh token failed err: %v", op, jwtErr)
		return Session{}, common.NewErrorWithDescription(SignupErrorInternal, jwtErr.Error())
	}
	transaction := c.authRepository.CreateUser(auth.UserId(uid), email, password, string(refreshToken))
	if err := transaction.Perform(); err != nil {
		log.Printf("storing credentials to db failed err: %v", err)
		return Session{}, common.NewErrorWithDescription(SignupErrorInternal, err.Error())
	}
	log.Printf("%s: success", op)
	return Session{
		Id:           UserId(uid),
		AccessToken:  string(accessToken),
		RefreshToken: string(refreshToken),
	}, nil
}

func (c *defaultController) Login(email string, password string) (Session, *common.CodeBasedError[LoginErrorCode]) {
	const op = "auth.defaultController.Login"
	log.Printf("%s: start", op)
	valid, err := c.authRepository.CheckCredentials(email, password)
	if err != nil {
		log.Printf("%s: credentials check failed err: %v", op, err)
		return Session{}, common.NewErrorWithDescription(LoginErrorInternal, err.Error())
	}
	if !valid {
		log.Printf("%s: credentials are wrong", op)
		return Session{}, common.NewError(LoginErrorWrongCredentials)
	}
	uid, err := c.authRepository.GetUserIdByEmail(email)
	if err != nil {
		log.Printf("%s: getting uid by credentials in db failed err: %v", op, err)
		return Session{}, common.NewErrorWithDescription(LoginErrorInternal, err.Error())
	}
	if uid == nil {
		log.Printf("%s: no uid accosiated with credentials", op)
		return Session{}, common.NewErrorWithDescription(LoginErrorInternal, "no uid accosiated with credentials")
	}
	accessToken, jwtErr := c.jwtService.IssueAccessToken(jwt.Subject(*uid))
	if jwtErr != nil {
		log.Printf("%s: issuing access token failed err: %v", op, jwtErr)
		return Session{}, common.NewErrorWithDescription(LoginErrorInternal, jwtErr.Error())
	}
	refreshToken, jwtErr := c.jwtService.IssueRefreshToken(jwt.Subject(*uid))
	if jwtErr != nil {
		log.Printf("%s: issuing refresh token failed err: %v", op, jwtErr)
		return Session{}, common.NewErrorWithDescription(LoginErrorInternal, jwtErr.Error())
	}
	transaction := c.authRepository.UpdateRefreshToken(*uid, string(refreshToken))
	if err := transaction.Perform(); err != nil {
		log.Printf("%s: storing refresh token to db failed err: %v", op, err)
		return Session{}, common.NewErrorWithDescription(LoginErrorInternal, err.Error())
	}
	log.Printf("%s: success", op)
	return Session{
		Id:           UserId(*uid),
		AccessToken:  string(accessToken),
		RefreshToken: string(refreshToken),
	}, nil
}

func (c *defaultController) Refresh(refreshToken string) (Session, *common.CodeBasedError[RefreshErrorCode]) {
	const op = "auth.defaultController.Refresh"
	log.Printf("%s: start", op)
	if err := c.jwtService.ValidateRefreshToken(jwt.RefreshToken(refreshToken)); err != nil {
		log.Printf("%s: token validation failed err: %v", op, err)
		switch err.Code {
		case jwt.CodeTokenExpired:
			return Session{}, common.NewErrorWithDescription(RefreshErrorTokenExpired, err.Error())
		case jwt.CodeTokenInvalid:
			return Session{}, common.NewErrorWithDescription(RefreshErrorTokenIsWrong, err.Error())
		default:
			return Session{}, common.NewErrorWithDescription(RefreshErrorInternal, err.Error())
		}
	}
	uid, err := c.jwtService.GetRefreshTokenSubject(jwt.RefreshToken(refreshToken))
	if err != nil {
		log.Printf("%s: cannot get refresh token subject err: %v", op, err)
		return Session{}, common.NewErrorWithDescription(RefreshErrorInternal, err.Error())
	}
	user, errGetFromDb := c.authRepository.GetUserInfo(auth.UserId(uid))
	if errGetFromDb != nil {
		log.Printf("%s: cannot get user data from db err: %v", op, errGetFromDb)
		return Session{}, common.NewErrorWithDescription(RefreshErrorInternal, errGetFromDb.Error())
	}
	if user.RefreshToken != refreshToken {
		log.Printf("%s: existed refresh token does not match with provided token", op)
		return Session{}, common.NewError(RefreshErrorTokenIsWrong)
	}
	newAccessToken, err := c.jwtService.IssueAccessToken(jwt.Subject(uid))
	if err != nil {
		log.Printf("%s: issuing access token failed err: %v", op, err)
		return Session{}, common.NewErrorWithDescription(RefreshErrorInternal, err.Error())
	}
	newRefreshToken, err := c.jwtService.IssueRefreshToken(jwt.Subject(uid))
	if err != nil {
		log.Printf("%s: issuing refresh token failed err: %v", op, err)
		return Session{}, common.NewErrorWithDescription(RefreshErrorInternal, err.Error())
	}
	transaction := c.authRepository.UpdateRefreshToken(auth.UserId(uid), refreshToken)
	if err := transaction.Perform(); err != nil {
		log.Printf("%s: storing refresh token to db failed err: %v", op, err)
		return Session{}, common.NewErrorWithDescription(RefreshErrorInternal, err.Error())
	}
	log.Printf("%s: success", op)
	return Session{
		Id:           UserId(uid),
		AccessToken:  string(newAccessToken),
		RefreshToken: string(newRefreshToken),
	}, nil
}

func (c *defaultController) Logout(id UserId) *common.CodeBasedError[LogoutErrorCode] {
	const op = "auth.defaultController.Logout"
	log.Printf("%s: start[id=%s]", op, id)
	refreshToken, jwtErr := c.jwtService.IssueRefreshToken(jwt.Subject(id))
	if jwtErr != nil {
		log.Printf("%s: issuing refresh token failed err: %v", op, jwtErr)
		return common.NewErrorWithDescription(LogoutErrorInternal, jwtErr.Error())
	}
	updateTokenTransaction := c.authRepository.UpdateRefreshToken(auth.UserId(id), string(refreshToken))
	if err := updateTokenTransaction.Perform(); err != nil {
		log.Printf("%s: storing refresh token to db failed err: %v", op, err)
		return common.NewErrorWithDescription(LogoutErrorInternal, err.Error())
	}
	log.Printf("%s: success[id=%s]", op, id)
	return nil
}

func (c *defaultController) UpdateEmail(email string, id UserId) (Session, *common.CodeBasedError[UpdateEmailErrorCode]) {
	const op = "auth.defaultController.UpdateEmail"
	log.Printf("%s: start[id=%s]", op, id)
	if err := c.formatValidationService.ValidateEmailFormat(email); err != nil {
		log.Printf("%s: wrong email format err: %v", op, err)
		return Session{}, common.NewErrorWithDescription(UpdateEmailErrorWrongFormat, err.Error())
	}
	uidForNewEmail, err := c.authRepository.GetUserIdByEmail(email)
	if err != nil {
		log.Printf("%s: cannot check email existence in db err: %v", op, err)
		return Session{}, common.NewErrorWithDescription(UpdateEmailErrorInternal, err.Error())
	}
	if uidForNewEmail != nil {
		log.Printf("%s: email is already taken", op)
		return Session{}, common.NewError(UpdateEmailErrorAlreadyTaken)
	}
	accessToken, jwtErr := c.jwtService.IssueAccessToken(jwt.Subject(id))
	if jwtErr != nil {
		log.Printf("%s: issuing access token failed err: %v", op, jwtErr)
		return Session{}, common.NewErrorWithDescription(UpdateEmailErrorInternal, jwtErr.Error())
	}
	refreshToken, jwtErr := c.jwtService.IssueRefreshToken(jwt.Subject(id))
	if jwtErr != nil {
		log.Printf("%s: issuing refresh token failed err: %v", op, jwtErr)
		return Session{}, common.NewErrorWithDescription(UpdateEmailErrorInternal, jwtErr.Error())
	}
	updateEmailTransaction := c.authRepository.UpdateEmail(auth.UserId(id), email)
	if err := updateEmailTransaction.Perform(); err != nil {
		log.Printf("%s: cannot update email in db err: %v", op, err)
		return Session{}, common.NewErrorWithDescription(UpdateEmailErrorInternal, err.Error())
	}
	updateTokenTransaction := c.authRepository.UpdateRefreshToken(auth.UserId(id), string(refreshToken))
	if err := updateTokenTransaction.Perform(); err != nil {
		log.Printf("%s: storing refresh token to db failed err: %v", op, err)
		updateEmailTransaction.Rollback()
		return Session{}, common.NewErrorWithDescription(UpdateEmailErrorInternal, err.Error())
	}
	log.Printf("%s: success[id=%s]", op, id)
	return Session{
		Id:           id,
		AccessToken:  string(accessToken),
		RefreshToken: string(refreshToken),
	}, nil
}

func (c *defaultController) UpdatePassword(oldPassword string, newPassword string, id UserId) (Session, *common.CodeBasedError[UpdatePasswordErrorCode]) {
	const op = "auth.defaultController.UpdatePassword"
	log.Printf("%s: start[id=%s]", op, id)
	if err := c.formatValidationService.ValidatePasswordFormat(newPassword); err != nil {
		log.Printf("%s: wrong password format err: %v", op, err)
		return Session{}, common.NewErrorWithDescription(UpdatePasswordErrorWrongFormat, err.Error())
	}
	account, err := c.authRepository.GetUserInfo(auth.UserId(id))
	if err != nil {
		log.Printf("%s: cannot get credentials for id in db err: %v", op, err)
		return Session{}, common.NewErrorWithDescription(UpdatePasswordErrorInternal, err.Error())
	}
	passed, err := c.authRepository.CheckCredentials(account.Email, oldPassword)
	if err != nil {
		log.Printf("%s: cannot check password for id in db err: %v", op, err)
		return Session{}, common.NewErrorWithDescription(UpdatePasswordErrorInternal, err.Error())
	}
	if !passed {
		log.Printf("%s: old password is wrong", op)
		return Session{}, common.NewError(UpdatePasswordErrorOldPasswordIsWrong)
	}
	accessToken, jwtErr := c.jwtService.IssueAccessToken(jwt.Subject(id))
	if jwtErr != nil {
		log.Printf("%s: issuing access token failed err: %v", op, jwtErr)
		return Session{}, common.NewErrorWithDescription(UpdatePasswordErrorInternal, jwtErr.Error())
	}
	refreshToken, jwtErr := c.jwtService.IssueRefreshToken(jwt.Subject(id))
	if jwtErr != nil {
		log.Printf("%s: issuing refresh token failed err: %v", op, jwtErr)
		return Session{}, common.NewErrorWithDescription(UpdatePasswordErrorInternal, jwtErr.Error())
	}
	updatePasswordTransaction := c.authRepository.UpdatePassword(auth.UserId(id), newPassword)
	if err := updatePasswordTransaction.Perform(); err != nil {
		log.Printf("%s: cannot update password in db err: %v", op, err)
		return Session{}, common.NewErrorWithDescription(UpdatePasswordErrorInternal, err.Error())
	}
	updateTokenTransaction := c.authRepository.UpdateRefreshToken(auth.UserId(id), string(refreshToken))
	if err := updateTokenTransaction.Perform(); err != nil {
		log.Printf("%s: storing refresh token to db failed err: %v", op, err)
		updatePasswordTransaction.Rollback()
		return Session{}, common.NewErrorWithDescription(UpdatePasswordErrorInternal, err.Error())
	}
	log.Printf("%s: success[id=%s]", op, id)
	return Session{
		Id:           id,
		AccessToken:  string(accessToken),
		RefreshToken: string(refreshToken),
	}, nil
}

func (c *defaultController) RegisterForPushNotifications(pushToken string, id UserId) *common.CodeBasedError[RegisterForPushNotificationsErrorCode] {
	const op = "auth.defaultController.ConfirmEmail"
	log.Printf("%s: start[id=%s]", op, id)
	storeTransaction := c.pushTokensRepository.StorePushToken(pushNotifications.UserId(id), pushToken)
	if err := storeTransaction.Perform(); err != nil {
		log.Printf("%s: cannot store push token in db err: %v", op, err)
		return common.NewErrorWithDescription(RegisterForPushNotificationsErrorInternal, err.Error())
	}
	log.Printf("%s: success[id=%s]", op, id)
	return nil
}
