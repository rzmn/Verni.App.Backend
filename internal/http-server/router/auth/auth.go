package auth

import (
	"errors"
	"fmt"
	"log"
	"net/mail"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"accounty/internal/auth/confirmation"
	"accounty/internal/auth/jwt"
	"accounty/internal/storage"

	"accounty/internal/http-server/handlers/auth/confirmEmail"
	"accounty/internal/http-server/handlers/auth/login"
	"accounty/internal/http-server/handlers/auth/logout"
	"accounty/internal/http-server/handlers/auth/refresh"
	"accounty/internal/http-server/handlers/auth/registerForPushNotifications"
	"accounty/internal/http-server/handlers/auth/sendEmailConfirmationCode"
	"accounty/internal/http-server/handlers/auth/signup"
	"accounty/internal/http-server/handlers/auth/updateEmail"
	"accounty/internal/http-server/handlers/auth/updatePassword"
	"accounty/internal/http-server/middleware"
)

func validateUserCredentialsFormat(credentials storage.UserCredentials) error {
	if err := validateEmailFormat(credentials.Email); err != nil {
		return err
	}
	if len(credentials.Password) < 6 {
		return errors.New("`password` field should contain more than 6 characters")
	}
	return nil
}

func validateEmailFormat(email string) error {
	_, err := mail.ParseAddress(email)
	if err != nil {
		return fmt.Errorf("`email` field is invalid: %v", err)
	}
	if strings.TrimSpace(email) != email {
		return fmt.Errorf("`email` field is invalid: leading or trailing spaces")
	}
	return nil
}

type loginRequestHandler struct {
	storage    storage.Storage
	jwtService jwt.Service
}

func (h *loginRequestHandler) Validate(request login.Request) *login.Error {
	const op = "router.loginRequestHandler.Validate"

	log.Printf("%s: validating start", op)
	if err := validateUserCredentialsFormat(request.Credentials); err != nil {
		log.Printf("%s: format validating failed %v", op, err)
		outError := login.ErrWrongCredentialsFormat()
		return &outError
	}
	log.Printf("%s: validated format", op)
	log.Printf("%s: checking email/pwd pair matches", op)
	valid, err := h.storage.CheckCredentials(request.Credentials)
	if err != nil {
		log.Printf("%s: checking email/pwd pair matches failed %v", op, err)
		outError := login.ErrInternal()
		return &outError
	}
	log.Printf("%s: checked email/pwd pair matches without errors", op)
	if !valid {
		log.Printf("%s: email/pwd pair did not match", op)
		outError := login.ErrIncorrectCredentials()
		return &outError
	}
	log.Printf("%s: email/pwd pair matches", op)
	return nil
}

func (h *loginRequestHandler) Handle(request login.Request) (*storage.AuthenticatedSession, *login.Error) {
	const op = "router.loginRequestHandler.Handle"

	log.Printf("%s: getting uid", op)
	uid, err := h.storage.GetUserId(request.Credentials.Email)
	if err != nil {
		log.Printf("%s: getting uid failed: %v", op, err)
		outError := login.ErrInternal()
		return nil, &outError
	}
	if uid == nil {
		log.Printf("%s: no uid accosiated with email: %s", op, request.Credentials.Email)
		outError := login.ErrInternal()
		return nil, &outError
	}
	log.Printf("%s: issuing tokens", op)
	accessToken, err := h.jwtService.IssueAccessToken(jwt.Subject(*uid))
	if err != nil {
		log.Printf("%s: issuing access token failed %v", op, err)
		outError := login.ErrInternal()
		return nil, &outError
	}
	refreshToken, err := h.jwtService.IssueRefreshToken(jwt.Subject(*uid))
	if err != nil {
		log.Printf("%s: issuing refresh token failed %v", op, err)
		outError := login.ErrInternal()
		return nil, &outError
	}
	log.Printf("%s: issued tokens ok", op)
	log.Printf("%s: storing refresh token", op)
	if err := h.storage.StoreRefreshToken(string(refreshToken), *uid); err != nil {
		log.Printf("%s: storing refresh token failed %v", op, err)
		outError := login.ErrInternal()
		return nil, &outError
	}
	log.Printf("%s: storing refresh token ok", op)
	return &storage.AuthenticatedSession{
		Id:           *uid,
		AccessToken:  string(accessToken),
		RefreshToken: string(refreshToken),
	}, nil
}

type signupRequestHandler struct {
	storage    storage.Storage
	jwtService jwt.Service
}

func (h *signupRequestHandler) Validate(request signup.Request) *signup.Error {
	const op = "router.signupRequestHandler.Validate"
	log.Printf("%s: start with request %v", op, request)
	if err := validateUserCredentialsFormat(request.Credentials); err != nil {
		log.Printf("%s: format validating failed %v", op, err)
		outError := signup.ErrWrongCredentialsFormat()
		return &outError
	}
	log.Printf("%s: validated format", op)
	uid, err := h.storage.GetUserId(request.Credentials.Email)
	if err != nil {
		log.Printf("%s: check user exists failed %v", op, err)
		outError := signup.ErrInternal()
		return &outError
	}
	if uid != nil {
		log.Printf("%s: already taken", op)
		outError := signup.ErrLoginAlreadyTaken()
		return &outError
	}
	log.Printf("%s: existance check ok", op)
	return nil
}

func (h *signupRequestHandler) Handle(request signup.Request) (*storage.AuthenticatedSession, *signup.Error) {
	const op = "router.signupRequestHandler.Handle"
	log.Printf("%s: start with request %v", op, request)
	uid := storage.UserId(uuid.New().String())
	log.Printf("%s: created new user id %s", op, uid)
	if err := h.storage.StoreCredentials(uid, request.Credentials); err != nil {
		log.Printf("storing credentials failed %v", err)
		outError := signup.ErrInternal()
		return nil, &outError
	}
	log.Printf("%s: credentials stored", op)
	log.Printf("%s: issuing tokens", op)
	accessToken, err := h.jwtService.IssueAccessToken(jwt.Subject(uid))
	if err != nil {
		log.Printf("%s: issuing access token failed %v", op, err)
		outError := signup.ErrInternal()
		return nil, &outError
	}
	refreshToken, err := h.jwtService.IssueRefreshToken(jwt.Subject(uid))
	if err != nil {
		log.Printf("%s: issuing refresh token failed %v", op, err)
		outError := signup.ErrInternal()
		return nil, &outError
	}
	log.Printf("%s: issued tokens ok", op)
	log.Printf("%s: storing refresh token", op)
	if err := h.storage.StoreRefreshToken(string(refreshToken), uid); err != nil {
		log.Printf("store tokens failed %v", err)
		outError := signup.ErrInternal()
		return nil, &outError
	}
	log.Printf("%s: tokens stored", op)
	log.Printf("%s: ok", op)
	return &storage.AuthenticatedSession{
		Id:           uid,
		AccessToken:  string(accessToken),
		RefreshToken: string(refreshToken),
	}, nil
}

type refreshRequestHandler struct {
	storage    storage.Storage
	jwtService jwt.Service
}

func (h *refreshRequestHandler) Handle(request refresh.Request) (*storage.AuthenticatedSession, *refresh.Error) {
	const op = "router.refreshRequestHandler.Handle"
	log.Printf("%s: start with request %v", op, request)
	if err := h.jwtService.ValidateRefreshToken(jwt.RefreshToken(request.RefreshToken)); err != nil {
		switch err.Code {
		case jwt.CodeTokenExpired:
			outError := refresh.ErrTokenExpired()
			return nil, &outError
		case jwt.CodeTokenInvalid:
			outError := refresh.ErrWrongAccessToken()
			return nil, &outError
		default:
			outError := refresh.ErrInternal()
			return nil, &outError
		}
	}
	subject, err := h.jwtService.GetRefreshTokenSubject(jwt.RefreshToken(request.RefreshToken))
	if err != nil {
		outError := refresh.ErrInternal()
		return nil, &outError
	}
	tokenFromDb, errGetFromDb := h.storage.GetRefreshToken(storage.UserId(subject))
	if errGetFromDb != nil {
		outError := refresh.ErrInternal()
		return nil, &outError
	}
	if tokenFromDb != nil && *tokenFromDb != request.RefreshToken {
		outError := refresh.ErrWrongAccessToken()
		return nil, &outError
	}
	log.Printf("%s: issuing tokens", op)
	accessToken, err := h.jwtService.IssueAccessToken(subject)
	if err != nil {
		log.Printf("%s: issuing access token failed %v", op, err)
		outError := refresh.ErrInternal()
		return nil, &outError
	}
	refreshToken, err := h.jwtService.IssueRefreshToken(subject)
	if err != nil {
		log.Printf("%s: issuing refresh token failed %v", op, err)
		outError := refresh.ErrInternal()
		return nil, &outError
	}
	log.Printf("%s: issued tokens ok", op)
	log.Printf("%s: storing refresh token", op)
	if err := h.storage.StoreRefreshToken(string(refreshToken), storage.UserId(subject)); err != nil {
		log.Printf("store tokens failed %v", err)
		outError := refresh.ErrInternal()
		return nil, &outError
	}
	log.Printf("%s: tokens stored", op)
	log.Printf("%s: ok", op)
	return &storage.AuthenticatedSession{
		Id:           storage.UserId(subject),
		AccessToken:  string(accessToken),
		RefreshToken: string(refreshToken),
	}, nil
}

type updateEmailRequestHandler struct {
	storage    storage.Storage
	jwtService jwt.Service
}

func (h *updateEmailRequestHandler) Validate(c *gin.Context, request updateEmail.Request) *updateEmail.Error {
	const op = "router.auth.updateEmailRequestHandler.Validate"
	if err := validateEmailFormat(request.Email); err != nil {
		outError := updateEmail.ErrWrongFormat()
		return &outError
	}
	exists, err := h.storage.IsEmailExists(request.Email)
	if err != nil {
		log.Printf("%s: cannot check email existence %v", op, err)
		outError := updateEmail.ErrInternal()
		return &outError
	}
	if exists {
		log.Printf("%s: email already taken %v", op, err)
		outError := updateEmail.ErrCodeAlreadyTaken()
		return &outError
	}
	return nil
}

func (h *updateEmailRequestHandler) Handle(c *gin.Context, request updateEmail.Request) (storage.AuthenticatedSession, *updateEmail.Error) {
	const op = "router.auth.updateEmailRequestHandler.Handle"
	uid := storage.UserId(c.Request.Header.Get(middleware.LoggedInSubjectKey))

	if err := h.storage.UpdateEmail(uid, request.Email); err != nil {
		log.Printf("%s: cannot update email %v", op, err)
		outError := updateEmail.ErrInternal()
		return storage.AuthenticatedSession{}, &outError
	}
	log.Printf("%s: issuing tokens", op)
	accessToken, err := h.jwtService.IssueAccessToken(jwt.Subject(uid))
	if err != nil {
		log.Printf("%s: issuing access token failed %v", op, err)
		outError := updateEmail.ErrInternal()
		return storage.AuthenticatedSession{}, &outError
	}
	refreshToken, err := h.jwtService.IssueRefreshToken(jwt.Subject(uid))
	if err != nil {
		log.Printf("%s: issuing refresh token failed %v", op, err)
		outError := updateEmail.ErrInternal()
		return storage.AuthenticatedSession{}, &outError
	}
	log.Printf("%s: issued tokens ok", op)
	log.Printf("%s: storing refresh token", op)
	if err := h.storage.StoreRefreshToken(string(refreshToken), uid); err != nil {
		log.Printf("%s: storing refresh token failed %v", op, err)
		outError := updateEmail.ErrInternal()
		return storage.AuthenticatedSession{}, &outError
	}
	log.Printf("%s: storing refresh token ok", op)
	return storage.AuthenticatedSession{
		Id:           uid,
		AccessToken:  string(accessToken),
		RefreshToken: string(refreshToken),
	}, nil
}

type updatePasswordRequestHandler struct {
	storage    storage.Storage
	jwtService jwt.Service
}

func (h *updatePasswordRequestHandler) Validate(c *gin.Context, request updatePassword.Request) *updatePassword.Error {
	const op = "router.auth.updatePasswordRequestHandler.Validate"
	uid := storage.UserId(c.Request.Header.Get(middleware.LoggedInSubjectKey))
	passed, err := h.storage.CheckPasswordForId(uid, request.OldPassword)
	if err != nil {
		log.Printf("%s: cannot check password for id %v", op, err)
		outError := updatePassword.ErrInternal()
		return &outError
	}
	if !passed {
		log.Printf("%s: password did not match", op)
		outError := updatePassword.ErrIncorrectCredentials()
		return &outError
	}
	return nil
}

func (h *updatePasswordRequestHandler) Handle(c *gin.Context, request updatePassword.Request) (storage.AuthenticatedSession, *updatePassword.Error) {
	const op = "router.auth.updatePasswordRequestHandler.Handle"
	uid := storage.UserId(c.Request.Header.Get(middleware.LoggedInSubjectKey))
	if err := h.storage.UpdatePasswordForId(uid, request.NewPassword); err != nil {
		log.Printf("%s: cannot update password: %v", op, err)
		outError := updatePassword.ErrInternal()
		return storage.AuthenticatedSession{}, &outError
	}
	log.Printf("%s: issuing tokens", op)
	accessToken, err := h.jwtService.IssueAccessToken(jwt.Subject(uid))
	if err != nil {
		log.Printf("%s: issuing access token failed %v", op, err)
		outError := updatePassword.ErrInternal()
		return storage.AuthenticatedSession{}, &outError
	}
	refreshToken, err := h.jwtService.IssueRefreshToken(jwt.Subject(uid))
	if err != nil {
		log.Printf("%s: issuing refresh token failed %v", op, err)
		outError := updatePassword.ErrInternal()
		return storage.AuthenticatedSession{}, &outError
	}
	log.Printf("%s: issued tokens ok", op)
	log.Printf("%s: storing refresh token", op)
	if err := h.storage.StoreRefreshToken(string(refreshToken), uid); err != nil {
		log.Printf("%s: storing refresh token failed %v", op, err)
		outError := updatePassword.ErrInternal()
		return storage.AuthenticatedSession{}, &outError
	}
	log.Printf("%s: storing refresh token ok", op)
	return storage.AuthenticatedSession{
		Id:           uid,
		AccessToken:  string(accessToken),
		RefreshToken: string(refreshToken),
	}, nil
}

type confirmEmailRequestHandler struct {
	storage      storage.Storage
	confirmation confirmation.EmailConfirmation
}

func (h *confirmEmailRequestHandler) Handle(c *gin.Context, request confirmEmail.Request) *confirmEmail.Error {
	const op = "router.auth.confirmEmailRequestHandler.Handle"
	log.Printf("%s: start", op)
	uid := storage.UserId(c.Request.Header.Get(middleware.LoggedInSubjectKey))
	account, err := h.storage.GetAccountInfo(uid)
	if err != nil || account == nil {
		log.Printf("%s: cannot get account info %v", op, err)
		outError := confirmEmail.ErrInternal()
		return &outError
	}
	if account.EmailVerified {
		log.Printf("%s: email already verified", op)
		return nil
	}
	if err := h.confirmation.ConfirmEmail(account.Email, request.Code); err != nil {
		log.Printf("%s: confirmation failed: %v", op, err)
		if errors.Is(err, confirmation.ErrCodeDidNotMatch) {
			outError := confirmEmail.ErrIncorrect()
			return &outError
		} else {
			outError := confirmEmail.ErrInternal()
			return &outError
		}
	}
	return nil
}

type sendEmailConfirmationCodeRequestHandler struct {
	storage      storage.Storage
	confirmation confirmation.EmailConfirmation
}

func (h *sendEmailConfirmationCodeRequestHandler) Handle(c *gin.Context, request sendEmailConfirmationCode.Request) *sendEmailConfirmationCode.Error {
	const op = "router.auth.confirmEmailRequestHandler.Handle"
	log.Printf("%s: start", op)
	uid := storage.UserId(c.Request.Header.Get(middleware.LoggedInSubjectKey))
	account, err := h.storage.GetAccountInfo(uid)
	if err != nil || account == nil {
		log.Printf("%s: cannot get account info %v", op, err)
		outError := sendEmailConfirmationCode.ErrInternal()
		return &outError
	}
	if account.EmailVerified {
		log.Printf("%s: email already verified", op)
		outError := sendEmailConfirmationCode.ErrAlreadyConfirmed()
		return &outError
	}
	if err := h.confirmation.SendConfirmationCode(account.Email); err != nil {
		if errors.Is(err, confirmation.ErrNotDeliveded) {
			outError := sendEmailConfirmationCode.ErrNotDelivered()
			return &outError
		} else {
			outError := sendEmailConfirmationCode.ErrInternal()
			return &outError
		}
	}
	return nil
}

type registerForPushNotificationsRequestHandler struct {
	storage storage.Storage
}

func (h *registerForPushNotificationsRequestHandler) Handle(c *gin.Context, request registerForPushNotifications.Request) *registerForPushNotifications.Error {
	const op = "router.auth.registerForPushNotificationsRequestHandler.Handle"
	log.Printf("%s: start", op)
	uid := storage.UserId(c.Request.Header.Get(middleware.LoggedInSubjectKey))
	if err := h.storage.StorePushToken(uid, request.Token); err != nil {
		log.Printf("%s: cannot store push token %v", op, err)
		outError := registerForPushNotifications.ErrInternal()
		return &outError
	}
	return nil
}

type logoutRequestHandler struct {
	storage storage.Storage
}

func (h *logoutRequestHandler) Handle(c *gin.Context) *logout.Error {
	const op = "router.auth.logoutRequestHandler.Handle"
	log.Printf("%s: start", op)
	uid := storage.UserId(c.Request.Header.Get(middleware.LoggedInSubjectKey))
	if err := h.storage.RemoveRefreshToken(uid); err != nil {
		log.Printf("%s: cannot remove refresh token %v", op, err)
		outError := logout.ErrInternal()
		return &outError
	}
	return nil
}

func RegisterRoutes(e *gin.Engine, storage storage.Storage, jwtService jwt.Service) {
	ensureLogIn := middleware.EnsureLoggedIn(storage, jwtService)
	e.PUT(
		"/auth/signup",
		signup.New(
			&signupRequestHandler{
				storage: storage,
			},
		),
	)
	e.PUT(
		"/auth/login",
		login.New(
			&loginRequestHandler{
				storage: storage,
			},
		),
	)
	e.PUT(
		"/auth/refresh",
		refresh.New(
			&refreshRequestHandler{
				storage: storage,
			},
		),
	)
	e.PUT(
		"/auth/updateEmail",
		ensureLogIn,
		updateEmail.New(
			&updateEmailRequestHandler{
				storage:    storage,
				jwtService: jwtService,
			},
		),
	)
	e.PUT(
		"/auth/updatePassword",
		ensureLogIn,
		updatePassword.New(
			&updatePasswordRequestHandler{
				storage:    storage,
				jwtService: jwtService,
			},
		),
	)
	e.DELETE(
		"/auth/logout",
		ensureLogIn,
		logout.New(
			&logoutRequestHandler{
				storage: storage,
			},
		),
	)
	e.PUT(
		"/auth/confirmEmail",
		ensureLogIn,
		confirmEmail.New(
			&confirmEmailRequestHandler{
				storage: storage,
				confirmation: confirmation.EmailConfirmation{
					Storage: storage,
				},
			},
		),
	)
	e.PUT(
		"/auth/sendEmailConfirmationCode",
		ensureLogIn,
		sendEmailConfirmationCode.New(
			&sendEmailConfirmationCodeRequestHandler{
				storage: storage,
				confirmation: confirmation.EmailConfirmation{
					Storage: storage,
				},
			},
		),
	)
	e.PUT(
		"/auth/registerForPushNotifications",
		ensureLogIn,
		registerForPushNotifications.New(
			&registerForPushNotificationsRequestHandler{
				storage: storage,
			},
		),
	)
}
