package defaultController

import (
	"strings"

	"github.com/rzmn/governi/internal/common"

	"github.com/rzmn/governi/internal/services/formatValidation"
	"github.com/rzmn/governi/internal/services/jwt"
	"github.com/rzmn/governi/internal/services/logging"

	"github.com/rzmn/governi/internal/controllers/auth"

	authRepository "github.com/rzmn/governi/internal/repositories/auth"
	pushNotificationsRepository "github.com/rzmn/governi/internal/repositories/pushNotifications"
	usersRepository "github.com/rzmn/governi/internal/repositories/users"

	"github.com/google/uuid"
)

type AuthRepository authRepository.Repository
type UsersRepository usersRepository.Repository
type PushTokensRepository pushNotificationsRepository.Repository

func New(
	authRepository AuthRepository,
	pushTokensRepository PushTokensRepository,
	usersRepository UsersRepository,
	jwtService jwt.Service,
	formatValidationService formatValidation.Service,
	logger logging.Service,
) auth.Controller {
	return &defaultController{
		authRepository:          authRepository,
		pushTokensRepository:    pushTokensRepository,
		usersRepository:         usersRepository,
		jwtService:              jwtService,
		formatValidationService: formatValidationService,
		logger:                  logger,
	}
}

type defaultController struct {
	authRepository          AuthRepository
	pushTokensRepository    PushTokensRepository
	usersRepository         UsersRepository
	jwtService              jwt.Service
	formatValidationService formatValidation.Service
	logger                  logging.Service
}

func (c *defaultController) Signup(email string, password string) (auth.Session, *common.CodeBasedError[auth.SignupErrorCode]) {
	const op = "auth.defaultController.Signup"
	c.logger.LogInfo("%s: start", op)
	if err := c.formatValidationService.ValidateEmailFormat(email); err != nil {
		c.logger.LogInfo("%s: wrong email format err: %v", op, err)
		return auth.Session{}, common.NewErrorWithDescription(auth.SignupErrorWrongFormat, err.Error())
	}
	if err := c.formatValidationService.ValidatePasswordFormat(password); err != nil {
		c.logger.LogInfo("%s: wrong password format err: %v", op, err)
		return auth.Session{}, common.NewErrorWithDescription(auth.SignupErrorWrongFormat, err.Error())
	}
	uidAccosiatedWithEmail, err := c.authRepository.GetUserIdByEmail(email)
	if err != nil {
		c.logger.LogInfo("%s: getting uid by credentials from db failed err: %v", op, err)
		return auth.Session{}, common.NewErrorWithDescription(auth.SignupErrorInternal, err.Error())
	}
	if uidAccosiatedWithEmail != nil {
		c.logger.LogInfo("%s: already has an uid accosiated with credentials", op)
		return auth.Session{}, common.NewError(auth.SignupErrorAlreadyTaken)
	}
	uid := uuid.New().String()
	accessToken, jwtErr := c.jwtService.IssueAccessToken(jwt.Subject(uid))
	if jwtErr != nil {
		c.logger.LogInfo("%s: issuing access token failed err: %v", op, jwtErr)
		return auth.Session{}, common.NewErrorWithDescription(auth.SignupErrorInternal, jwtErr.Error())
	}
	refreshToken, jwtErr := c.jwtService.IssueRefreshToken(jwt.Subject(uid))
	if jwtErr != nil {
		c.logger.LogInfo("%s: issuing refresh token failed err: %v", op, jwtErr)
		return auth.Session{}, common.NewErrorWithDescription(auth.SignupErrorInternal, jwtErr.Error())
	}
	createUserTransaction := c.usersRepository.StoreUser(usersRepository.User{
		Id:          usersRepository.UserId(uid),
		DisplayName: strings.Split(email, "@")[0],
		AvatarId:    nil,
	})
	if err := createUserTransaction.Perform(); err != nil {
		c.logger.LogInfo("storing user meta to db failed err: %v", err)
		return auth.Session{}, common.NewErrorWithDescription(auth.SignupErrorInternal, err.Error())
	}
	transaction := c.authRepository.CreateUser(authRepository.UserId(uid), email, password, string(refreshToken))
	if err := transaction.Perform(); err != nil {
		createUserTransaction.Rollback()
		c.logger.LogInfo("storing credentials to db failed err: %v", err)
		return auth.Session{}, common.NewErrorWithDescription(auth.SignupErrorInternal, err.Error())
	}
	c.logger.LogInfo("%s: success", op)
	return auth.Session{
		Id:           auth.UserId(uid),
		AccessToken:  string(accessToken),
		RefreshToken: string(refreshToken),
	}, nil
}

func (c *defaultController) Login(email string, password string) (auth.Session, *common.CodeBasedError[auth.LoginErrorCode]) {
	const op = "auth.defaultController.Login"
	c.logger.LogInfo("%s: start", op)
	valid, err := c.authRepository.CheckCredentials(email, password)
	if err != nil {
		c.logger.LogInfo("%s: credentials check failed err: %v", op, err)
		return auth.Session{}, common.NewErrorWithDescription(auth.LoginErrorInternal, err.Error())
	}
	if !valid {
		c.logger.LogInfo("%s: credentials are wrong", op)
		return auth.Session{}, common.NewError(auth.LoginErrorWrongCredentials)
	}
	uid, err := c.authRepository.GetUserIdByEmail(email)
	if err != nil {
		c.logger.LogInfo("%s: getting uid by credentials in db failed err: %v", op, err)
		return auth.Session{}, common.NewErrorWithDescription(auth.LoginErrorInternal, err.Error())
	}
	if uid == nil {
		c.logger.LogInfo("%s: no uid accosiated with credentials", op)
		return auth.Session{}, common.NewErrorWithDescription(auth.LoginErrorInternal, "no uid accosiated with credentials")
	}
	accessToken, jwtErr := c.jwtService.IssueAccessToken(jwt.Subject(*uid))
	if jwtErr != nil {
		c.logger.LogInfo("%s: issuing access token failed err: %v", op, jwtErr)
		return auth.Session{}, common.NewErrorWithDescription(auth.LoginErrorInternal, jwtErr.Error())
	}
	refreshToken, jwtErr := c.jwtService.IssueRefreshToken(jwt.Subject(*uid))
	if jwtErr != nil {
		c.logger.LogInfo("%s: issuing refresh token failed err: %v", op, jwtErr)
		return auth.Session{}, common.NewErrorWithDescription(auth.LoginErrorInternal, jwtErr.Error())
	}
	transaction := c.authRepository.UpdateRefreshToken(*uid, string(refreshToken))
	if err := transaction.Perform(); err != nil {
		c.logger.LogInfo("%s: storing refresh token to db failed err: %v", op, err)
		return auth.Session{}, common.NewErrorWithDescription(auth.LoginErrorInternal, err.Error())
	}
	c.logger.LogInfo("%s: success", op)
	return auth.Session{
		Id:           auth.UserId(*uid),
		AccessToken:  string(accessToken),
		RefreshToken: string(refreshToken),
	}, nil
}

func (c *defaultController) Refresh(refreshToken string) (auth.Session, *common.CodeBasedError[auth.RefreshErrorCode]) {
	const op = "auth.defaultController.Refresh"
	c.logger.LogInfo("%s: start", op)
	if err := c.jwtService.ValidateRefreshToken(jwt.RefreshToken(refreshToken)); err != nil {
		c.logger.LogInfo("%s: token validation failed err: %v", op, err)
		switch err.Code {
		case jwt.CodeTokenExpired:
			return auth.Session{}, common.NewErrorWithDescription(auth.RefreshErrorTokenExpired, err.Error())
		case jwt.CodeTokenInvalid:
			return auth.Session{}, common.NewErrorWithDescription(auth.RefreshErrorTokenIsWrong, err.Error())
		default:
			return auth.Session{}, common.NewErrorWithDescription(auth.RefreshErrorInternal, err.Error())
		}
	}
	uid, err := c.jwtService.GetRefreshTokenSubject(jwt.RefreshToken(refreshToken))
	if err != nil {
		c.logger.LogInfo("%s: cannot get refresh token subject err: %v", op, err)
		return auth.Session{}, common.NewErrorWithDescription(auth.RefreshErrorInternal, err.Error())
	}
	user, errGetFromDb := c.authRepository.GetUserInfo(authRepository.UserId(uid))
	if errGetFromDb != nil {
		c.logger.LogInfo("%s: cannot get user data from db err: %v", op, errGetFromDb)
		return auth.Session{}, common.NewErrorWithDescription(auth.RefreshErrorInternal, errGetFromDb.Error())
	}
	if user.RefreshToken != refreshToken {
		c.logger.LogInfo("%s: existed refresh token does not match with provided token", op)
		return auth.Session{}, common.NewError(auth.RefreshErrorTokenIsWrong)
	}
	newAccessToken, err := c.jwtService.IssueAccessToken(jwt.Subject(uid))
	if err != nil {
		c.logger.LogInfo("%s: issuing access token failed err: %v", op, err)
		return auth.Session{}, common.NewErrorWithDescription(auth.RefreshErrorInternal, err.Error())
	}
	newRefreshToken, err := c.jwtService.IssueRefreshToken(jwt.Subject(uid))
	if err != nil {
		c.logger.LogInfo("%s: issuing refresh token failed err: %v", op, err)
		return auth.Session{}, common.NewErrorWithDescription(auth.RefreshErrorInternal, err.Error())
	}
	transaction := c.authRepository.UpdateRefreshToken(authRepository.UserId(uid), string(newRefreshToken))
	if err := transaction.Perform(); err != nil {
		c.logger.LogInfo("%s: storing refresh token to db failed err: %v", op, err)
		return auth.Session{}, common.NewErrorWithDescription(auth.RefreshErrorInternal, err.Error())
	}
	c.logger.LogInfo("%s: success", op)
	return auth.Session{
		Id:           auth.UserId(uid),
		AccessToken:  string(newAccessToken),
		RefreshToken: string(newRefreshToken),
	}, nil
}

func (c *defaultController) Logout(id auth.UserId) *common.CodeBasedError[auth.LogoutErrorCode] {
	const op = "auth.defaultController.Logout"
	c.logger.LogInfo("%s: start[id=%s]", op, id)
	refreshToken, jwtErr := c.jwtService.IssueRefreshToken(jwt.Subject(id))
	if jwtErr != nil {
		c.logger.LogInfo("%s: issuing refresh token failed err: %v", op, jwtErr)
		return common.NewErrorWithDescription(auth.LogoutErrorInternal, jwtErr.Error())
	}
	updateTokenTransaction := c.authRepository.UpdateRefreshToken(authRepository.UserId(id), string(refreshToken))
	if err := updateTokenTransaction.Perform(); err != nil {
		c.logger.LogInfo("%s: storing refresh token to db failed err: %v", op, err)
		return common.NewErrorWithDescription(auth.LogoutErrorInternal, err.Error())
	}
	c.logger.LogInfo("%s: success[id=%s]", op, id)
	return nil
}

func (c *defaultController) UpdateEmail(email string, id auth.UserId) (auth.Session, *common.CodeBasedError[auth.UpdateEmailErrorCode]) {
	const op = "auth.defaultController.UpdateEmail"
	c.logger.LogInfo("%s: start[id=%s]", op, id)
	if err := c.formatValidationService.ValidateEmailFormat(email); err != nil {
		c.logger.LogInfo("%s: wrong email format err: %v", op, err)
		return auth.Session{}, common.NewErrorWithDescription(auth.UpdateEmailErrorWrongFormat, err.Error())
	}
	uidForNewEmail, err := c.authRepository.GetUserIdByEmail(email)
	if err != nil {
		c.logger.LogInfo("%s: cannot check email existence in db err: %v", op, err)
		return auth.Session{}, common.NewErrorWithDescription(auth.UpdateEmailErrorInternal, err.Error())
	}
	if uidForNewEmail != nil {
		c.logger.LogInfo("%s: email is already taken", op)
		return auth.Session{}, common.NewError(auth.UpdateEmailErrorAlreadyTaken)
	}
	accessToken, jwtErr := c.jwtService.IssueAccessToken(jwt.Subject(id))
	if jwtErr != nil {
		c.logger.LogInfo("%s: issuing access token failed err: %v", op, jwtErr)
		return auth.Session{}, common.NewErrorWithDescription(auth.UpdateEmailErrorInternal, jwtErr.Error())
	}
	refreshToken, jwtErr := c.jwtService.IssueRefreshToken(jwt.Subject(id))
	if jwtErr != nil {
		c.logger.LogInfo("%s: issuing refresh token failed err: %v", op, jwtErr)
		return auth.Session{}, common.NewErrorWithDescription(auth.UpdateEmailErrorInternal, jwtErr.Error())
	}
	updateEmailTransaction := c.authRepository.UpdateEmail(authRepository.UserId(id), email)
	if err := updateEmailTransaction.Perform(); err != nil {
		c.logger.LogInfo("%s: cannot update email in db err: %v", op, err)
		return auth.Session{}, common.NewErrorWithDescription(auth.UpdateEmailErrorInternal, err.Error())
	}
	updateTokenTransaction := c.authRepository.UpdateRefreshToken(authRepository.UserId(id), string(refreshToken))
	if err := updateTokenTransaction.Perform(); err != nil {
		c.logger.LogInfo("%s: storing refresh token to db failed err: %v", op, err)
		updateEmailTransaction.Rollback()
		return auth.Session{}, common.NewErrorWithDescription(auth.UpdateEmailErrorInternal, err.Error())
	}
	c.logger.LogInfo("%s: success[id=%s]", op, id)
	return auth.Session{
		Id:           id,
		AccessToken:  string(accessToken),
		RefreshToken: string(refreshToken),
	}, nil
}

func (c *defaultController) UpdatePassword(oldPassword string, newPassword string, id auth.UserId) (auth.Session, *common.CodeBasedError[auth.UpdatePasswordErrorCode]) {
	const op = "auth.defaultController.UpdatePassword"
	c.logger.LogInfo("%s: start[id=%s]", op, id)
	if err := c.formatValidationService.ValidatePasswordFormat(newPassword); err != nil {
		c.logger.LogInfo("%s: wrong password format err: %v", op, err)
		return auth.Session{}, common.NewErrorWithDescription(auth.UpdatePasswordErrorWrongFormat, err.Error())
	}
	account, err := c.authRepository.GetUserInfo(authRepository.UserId(id))
	if err != nil {
		c.logger.LogInfo("%s: cannot get credentials for id in db err: %v", op, err)
		return auth.Session{}, common.NewErrorWithDescription(auth.UpdatePasswordErrorInternal, err.Error())
	}
	passed, err := c.authRepository.CheckCredentials(account.Email, oldPassword)
	if err != nil {
		c.logger.LogInfo("%s: cannot check password for id in db err: %v", op, err)
		return auth.Session{}, common.NewErrorWithDescription(auth.UpdatePasswordErrorInternal, err.Error())
	}
	if !passed {
		c.logger.LogInfo("%s: old password is wrong", op)
		return auth.Session{}, common.NewError(auth.UpdatePasswordErrorOldPasswordIsWrong)
	}
	accessToken, jwtErr := c.jwtService.IssueAccessToken(jwt.Subject(id))
	if jwtErr != nil {
		c.logger.LogInfo("%s: issuing access token failed err: %v", op, jwtErr)
		return auth.Session{}, common.NewErrorWithDescription(auth.UpdatePasswordErrorInternal, jwtErr.Error())
	}
	refreshToken, jwtErr := c.jwtService.IssueRefreshToken(jwt.Subject(id))
	if jwtErr != nil {
		c.logger.LogInfo("%s: issuing refresh token failed err: %v", op, jwtErr)
		return auth.Session{}, common.NewErrorWithDescription(auth.UpdatePasswordErrorInternal, jwtErr.Error())
	}
	updatePasswordTransaction := c.authRepository.UpdatePassword(authRepository.UserId(id), newPassword)
	if err := updatePasswordTransaction.Perform(); err != nil {
		c.logger.LogInfo("%s: cannot update password in db err: %v", op, err)
		return auth.Session{}, common.NewErrorWithDescription(auth.UpdatePasswordErrorInternal, err.Error())
	}
	updateTokenTransaction := c.authRepository.UpdateRefreshToken(authRepository.UserId(id), string(refreshToken))
	if err := updateTokenTransaction.Perform(); err != nil {
		c.logger.LogInfo("%s: storing refresh token to db failed err: %v", op, err)
		updatePasswordTransaction.Rollback()
		return auth.Session{}, common.NewErrorWithDescription(auth.UpdatePasswordErrorInternal, err.Error())
	}
	c.logger.LogInfo("%s: success[id=%s]", op, id)
	return auth.Session{
		Id:           id,
		AccessToken:  string(accessToken),
		RefreshToken: string(refreshToken),
	}, nil
}

func (c *defaultController) RegisterForPushNotifications(pushToken string, id auth.UserId) *common.CodeBasedError[auth.RegisterForPushNotificationsErrorCode] {
	const op = "auth.defaultController.ConfirmEmail"
	c.logger.LogInfo("%s: start[id=%s]", op, id)
	storeTransaction := c.pushTokensRepository.StorePushToken(pushNotificationsRepository.UserId(id), pushToken)
	if err := storeTransaction.Perform(); err != nil {
		c.logger.LogInfo("%s: cannot store push token in db err: %v", op, err)
		return common.NewErrorWithDescription(auth.RegisterForPushNotificationsErrorInternal, err.Error())
	}
	c.logger.LogInfo("%s: success[id=%s]", op, id)
	return nil
}
