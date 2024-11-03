package auth_test

import (
	"errors"
	"testing"
	"verni/internal/controllers/auth"
	"verni/internal/repositories"
	authRepository "verni/internal/repositories/auth"
	auth_mock "verni/internal/repositories/auth/mock"
	"verni/internal/repositories/pushNotifications"
	pushNotifications_mock "verni/internal/repositories/pushNotifications/mock"
	"verni/internal/repositories/users"
	users_mock "verni/internal/repositories/users/mock"
	formatValidation_mock "verni/internal/services/formatValidation/mock"
	"verni/internal/services/jwt"
	jwt_mock "verni/internal/services/jwt/mock"
	"verni/internal/services/logging"

	"github.com/google/uuid"
)

func TestSignupInvalidEmailFormat(t *testing.T) {
	formatValidatorMock := formatValidation_mock.ServiceMock{
		ValidateEmailFormatImpl: func(email string) error {
			return errors.New("some error")
		},
		ValidatePasswordFormatImpl: func(password string) error {
			return nil
		},
	}
	authRepositoryMock := auth_mock.RepositoryMock{}
	pushTokensRepositoryMock := pushNotifications_mock.RepositoryMock{}
	usersRepositoryMock := users_mock.RepositoryMock{}
	jwtServiceMock := jwt_mock.ServiceMock{}
	controller := auth.DefaultController(
		&authRepositoryMock,
		&pushTokensRepositoryMock,
		&usersRepositoryMock,
		&jwtServiceMock,
		&formatValidatorMock,
		logging.TestService(),
	)
	_, err := controller.Signup(uuid.New().String(), uuid.New().String())
	if err == nil {
		t.Fatalf("err should not be nil")
	}
	if err.Code != auth.SignupErrorWrongFormat {
		t.Fatalf("err code should be `wrong format`, found %v", err)
	}
}

func TestSignupInvalidPasswordFormat(t *testing.T) {
	formatValidatorMock := formatValidation_mock.ServiceMock{
		ValidateEmailFormatImpl: func(email string) error {
			return nil
		},
		ValidatePasswordFormatImpl: func(password string) error {
			return errors.New("some error")
		},
	}
	authRepositoryMock := auth_mock.RepositoryMock{}
	pushTokensRepositoryMock := pushNotifications_mock.RepositoryMock{}
	usersRepositoryMock := users_mock.RepositoryMock{}
	jwtServiceMock := jwt_mock.ServiceMock{}
	controller := auth.DefaultController(
		&authRepositoryMock,
		&pushTokensRepositoryMock,
		&usersRepositoryMock,
		&jwtServiceMock,
		&formatValidatorMock,
		logging.TestService(),
	)
	_, err := controller.Signup(uuid.New().String(), uuid.New().String())
	if err == nil {
		t.Fatalf("err should not be nil")
	}
	if err.Code != auth.SignupErrorWrongFormat {
		t.Fatalf("err code should be `wrong format`, found %v", err)
	}
}

func TestSignupFailedToCheckIfEmailIsTaken(t *testing.T) {
	formatValidatorMock := formatValidation_mock.ServiceMock{
		ValidateEmailFormatImpl: func(email string) error {
			return nil
		},
		ValidatePasswordFormatImpl: func(password string) error {
			return nil
		},
	}
	authRepositoryMock := auth_mock.RepositoryMock{
		GetUserIdByEmailImpl: func(email string) (*authRepository.UserId, error) {
			return nil, errors.New("some error")
		},
	}
	pushTokensRepositoryMock := pushNotifications_mock.RepositoryMock{}
	usersRepositoryMock := users_mock.RepositoryMock{}
	jwtServiceMock := jwt_mock.ServiceMock{}
	controller := auth.DefaultController(
		&authRepositoryMock,
		&pushTokensRepositoryMock,
		&usersRepositoryMock,
		&jwtServiceMock,
		&formatValidatorMock,
		logging.TestService(),
	)
	_, err := controller.Signup(uuid.New().String(), uuid.New().String())
	if err == nil {
		t.Fatalf("err should not be nil")
	}
	if err.Code != auth.SignupErrorInternal {
		t.Fatalf("err code should be `internal`, found %v", err)
	}
}

func TestSignupFailedEmailIsTaken(t *testing.T) {
	formatValidatorMock := formatValidation_mock.ServiceMock{
		ValidateEmailFormatImpl: func(email string) error {
			return nil
		},
		ValidatePasswordFormatImpl: func(password string) error {
			return nil
		},
	}
	authRepositoryMock := auth_mock.RepositoryMock{
		GetUserIdByEmailImpl: func(email string) (*authRepository.UserId, error) {
			id := uuid.New().String()
			return (*authRepository.UserId)(&id), nil
		},
	}
	pushTokensRepositoryMock := pushNotifications_mock.RepositoryMock{}
	usersRepositoryMock := users_mock.RepositoryMock{}
	jwtServiceMock := jwt_mock.ServiceMock{}
	controller := auth.DefaultController(
		&authRepositoryMock,
		&pushTokensRepositoryMock,
		&usersRepositoryMock,
		&jwtServiceMock,
		&formatValidatorMock,
		logging.TestService(),
	)
	_, err := controller.Signup(uuid.New().String(), uuid.New().String())
	if err == nil {
		t.Fatalf("err should not be nil")
	}
	if err.Code != auth.SignupErrorAlreadyTaken {
		t.Fatalf("err code should be `already taken`, found %v", err)
	}
}

func TestSignupIssueAccessTokenFailed(t *testing.T) {
	formatValidatorMock := formatValidation_mock.ServiceMock{
		ValidateEmailFormatImpl: func(email string) error {
			return nil
		},
		ValidatePasswordFormatImpl: func(password string) error {
			return nil
		},
	}
	authRepositoryMock := auth_mock.RepositoryMock{
		GetUserIdByEmailImpl: func(email string) (*authRepository.UserId, error) {
			return nil, nil
		},
	}
	pushTokensRepositoryMock := pushNotifications_mock.RepositoryMock{}
	usersRepositoryMock := users_mock.RepositoryMock{}
	jwtServiceMock := jwt_mock.ServiceMock{
		IssueAccessTokenImpl: func(subject jwt.Subject) (jwt.AccessToken, *jwt.Error) {
			return jwt.AccessToken(uuid.New().String()), &jwt.Error{}
		},
		IssueRefreshTokenImpl: func(subject jwt.Subject) (jwt.RefreshToken, *jwt.Error) {
			return jwt.RefreshToken(uuid.New().String()), nil
		},
	}
	controller := auth.DefaultController(
		&authRepositoryMock,
		&pushTokensRepositoryMock,
		&usersRepositoryMock,
		&jwtServiceMock,
		&formatValidatorMock,
		logging.TestService(),
	)
	_, err := controller.Signup(uuid.New().String(), uuid.New().String())
	if err == nil {
		t.Fatalf("err should not be nil")
	}
	if err.Code != auth.SignupErrorInternal {
		t.Fatalf("err code should be `internal`, found %v", err)
	}
}

func TestSignupIssueRefreshTokenFailed(t *testing.T) {
	formatValidatorMock := formatValidation_mock.ServiceMock{
		ValidateEmailFormatImpl: func(email string) error {
			return nil
		},
		ValidatePasswordFormatImpl: func(password string) error {
			return nil
		},
	}
	authRepositoryMock := auth_mock.RepositoryMock{
		GetUserIdByEmailImpl: func(email string) (*authRepository.UserId, error) {
			return nil, nil
		},
	}
	pushTokensRepositoryMock := pushNotifications_mock.RepositoryMock{}
	usersRepositoryMock := users_mock.RepositoryMock{}
	jwtServiceMock := jwt_mock.ServiceMock{
		IssueAccessTokenImpl: func(subject jwt.Subject) (jwt.AccessToken, *jwt.Error) {
			return jwt.AccessToken(uuid.New().String()), nil
		},
		IssueRefreshTokenImpl: func(subject jwt.Subject) (jwt.RefreshToken, *jwt.Error) {
			return jwt.RefreshToken(uuid.New().String()), &jwt.Error{}
		},
	}
	controller := auth.DefaultController(
		&authRepositoryMock,
		&pushTokensRepositoryMock,
		&usersRepositoryMock,
		&jwtServiceMock,
		&formatValidatorMock,
		logging.TestService(),
	)
	_, err := controller.Signup(uuid.New().String(), uuid.New().String())
	if err == nil {
		t.Fatalf("err should not be nil")
	}
	if err.Code != auth.SignupErrorInternal {
		t.Fatalf("err code should be `internal`, found %v", err)
	}
}

func TestSignupCreateUserMetaFailed(t *testing.T) {
	formatValidatorMock := formatValidation_mock.ServiceMock{
		ValidateEmailFormatImpl: func(email string) error {
			return nil
		},
		ValidatePasswordFormatImpl: func(password string) error {
			return nil
		},
	}
	authRepositoryMock := auth_mock.RepositoryMock{
		GetUserIdByEmailImpl: func(email string) (*authRepository.UserId, error) {
			return nil, nil
		},
	}
	pushTokensRepositoryMock := pushNotifications_mock.RepositoryMock{}
	usersRepositoryMock := users_mock.RepositoryMock{
		StoreUserImpl: func(user users.User) repositories.MutationWorkItem {
			return repositories.MutationWorkItem{
				Perform: func() error {
					return errors.New("some error")
				},
			}
		},
	}
	jwtServiceMock := jwt_mock.ServiceMock{
		IssueAccessTokenImpl: func(subject jwt.Subject) (jwt.AccessToken, *jwt.Error) {
			return jwt.AccessToken(uuid.New().String()), nil
		},
		IssueRefreshTokenImpl: func(subject jwt.Subject) (jwt.RefreshToken, *jwt.Error) {
			return jwt.RefreshToken(uuid.New().String()), nil
		},
	}
	controller := auth.DefaultController(
		&authRepositoryMock,
		&pushTokensRepositoryMock,
		&usersRepositoryMock,
		&jwtServiceMock,
		&formatValidatorMock,
		logging.TestService(),
	)
	_, err := controller.Signup(uuid.New().String(), uuid.New().String())
	if err == nil {
		t.Fatalf("err should not be nil")
	}
	if err.Code != auth.SignupErrorInternal {
		t.Fatalf("err code should be `internal`, found %v", err)
	}
}

func TestSignupCreateUserFailed(t *testing.T) {
	formatValidatorMock := formatValidation_mock.ServiceMock{
		ValidateEmailFormatImpl: func(email string) error {
			return nil
		},
		ValidatePasswordFormatImpl: func(password string) error {
			return nil
		},
	}
	authRepositoryMock := auth_mock.RepositoryMock{
		GetUserIdByEmailImpl: func(email string) (*authRepository.UserId, error) {
			return nil, nil
		},
		CreateUserImpl: func(uid authRepository.UserId, email, password, refreshToken string) repositories.MutationWorkItem {
			return repositories.MutationWorkItem{
				Perform: func() error {
					return errors.New("some error")
				},
			}
		},
	}
	pushTokensRepositoryMock := pushNotifications_mock.RepositoryMock{}
	storeUserCalls := 0
	storeUserRollbacks := 0
	usersRepositoryMock := users_mock.RepositoryMock{
		StoreUserImpl: func(user users.User) repositories.MutationWorkItem {
			return repositories.MutationWorkItem{
				Perform: func() error {
					storeUserCalls += 1
					return nil
				},
				Rollback: func() error {
					storeUserRollbacks += 1
					return nil
				},
			}
		},
	}
	jwtServiceMock := jwt_mock.ServiceMock{
		IssueAccessTokenImpl: func(subject jwt.Subject) (jwt.AccessToken, *jwt.Error) {
			return jwt.AccessToken(uuid.New().String()), nil
		},
		IssueRefreshTokenImpl: func(subject jwt.Subject) (jwt.RefreshToken, *jwt.Error) {
			return jwt.RefreshToken(uuid.New().String()), nil
		},
	}
	controller := auth.DefaultController(
		&authRepositoryMock,
		&pushTokensRepositoryMock,
		&usersRepositoryMock,
		&jwtServiceMock,
		&formatValidatorMock,
		logging.TestService(),
	)
	_, err := controller.Signup(uuid.New().String(), uuid.New().String())
	if err == nil {
		t.Fatalf("err should not be nil")
	}
	if err.Code != auth.SignupErrorInternal {
		t.Fatalf("err code should be `internal`, found %v", err)
	}
	if storeUserCalls != 1 || storeUserRollbacks != 1 {
		t.Fatalf("store and rollback should be called once, found %d %d", storeUserCalls, storeUserRollbacks)
	}
}

func TestSignupOk(t *testing.T) {
	createUserCalls := 0
	storeUserCalls := 0
	formatValidatorMock := formatValidation_mock.ServiceMock{
		ValidateEmailFormatImpl: func(email string) error {
			return nil
		},
		ValidatePasswordFormatImpl: func(password string) error {
			return nil
		},
	}
	authRepositoryMock := auth_mock.RepositoryMock{
		GetUserIdByEmailImpl: func(email string) (*authRepository.UserId, error) {
			return nil, nil
		},
		CreateUserImpl: func(uid authRepository.UserId, email, password, refreshToken string) repositories.MutationWorkItem {
			return repositories.MutationWorkItem{
				Perform: func() error {
					createUserCalls += 1
					return nil
				},
			}
		},
	}
	usersRepositoryMock := users_mock.RepositoryMock{
		StoreUserImpl: func(user users.User) repositories.MutationWorkItem {
			return repositories.MutationWorkItem{
				Perform: func() error {
					storeUserCalls += 1
					return nil
				},
			}
		},
	}
	pushTokensRepositoryMock := pushNotifications_mock.RepositoryMock{}
	jwtServiceMock := jwt_mock.ServiceMock{
		IssueAccessTokenImpl: func(subject jwt.Subject) (jwt.AccessToken, *jwt.Error) {
			return jwt.AccessToken(uuid.New().String()), nil
		},
		IssueRefreshTokenImpl: func(subject jwt.Subject) (jwt.RefreshToken, *jwt.Error) {
			return jwt.RefreshToken(uuid.New().String()), nil
		},
	}
	controller := auth.DefaultController(
		&authRepositoryMock,
		&pushTokensRepositoryMock,
		&usersRepositoryMock,
		&jwtServiceMock,
		&formatValidatorMock,
		logging.TestService(),
	)
	_, err := controller.Signup(uuid.New().String(), uuid.New().String())
	if err != nil {
		t.Fatalf("err should be nil, found %v", err)
	}
	if createUserCalls != 1 || storeUserCalls != 1 {
		t.Fatalf("`createUser` and `storeUser`, should be called once, found %d %d", createUserCalls, storeUserCalls)
	}
}

func TestLoginUnableToCheckCredentials(t *testing.T) {
	formatValidatorMock := formatValidation_mock.ServiceMock{}
	authRepositoryMock := auth_mock.RepositoryMock{
		CheckCredentialsImpl: func(email, password string) (bool, error) {
			return false, errors.New("some error")
		},
	}
	pushTokensRepositoryMock := pushNotifications_mock.RepositoryMock{}
	usersRepositoryMock := users_mock.RepositoryMock{}
	jwtServiceMock := jwt_mock.ServiceMock{}
	controller := auth.DefaultController(
		&authRepositoryMock,
		&pushTokensRepositoryMock,
		&usersRepositoryMock,
		&jwtServiceMock,
		&formatValidatorMock,
		logging.TestService(),
	)
	_, err := controller.Login(uuid.New().String(), uuid.New().String())
	if err == nil {
		t.Fatalf("err should not be nil")
	}
	if err.Code != auth.LoginErrorInternal {
		t.Fatalf("err code should be `internal`, found %v", err)
	}
}

func TestLoginWrongCredentials(t *testing.T) {
	formatValidatorMock := formatValidation_mock.ServiceMock{}
	authRepositoryMock := auth_mock.RepositoryMock{
		CheckCredentialsImpl: func(email, password string) (bool, error) {
			return false, nil
		},
	}
	pushTokensRepositoryMock := pushNotifications_mock.RepositoryMock{}
	usersRepositoryMock := users_mock.RepositoryMock{}
	jwtServiceMock := jwt_mock.ServiceMock{}
	controller := auth.DefaultController(
		&authRepositoryMock,
		&pushTokensRepositoryMock,
		&usersRepositoryMock,
		&jwtServiceMock,
		&formatValidatorMock,
		logging.TestService(),
	)
	_, err := controller.Login(uuid.New().String(), uuid.New().String())
	if err == nil {
		t.Fatalf("err should not be nil")
	}
	if err.Code != auth.LoginErrorWrongCredentials {
		t.Fatalf("err code should be `wrong credentials`, found %v", err)
	}
}

func TestLoginGetUserFailed(t *testing.T) {
	formatValidatorMock := formatValidation_mock.ServiceMock{}
	authRepositoryMock := auth_mock.RepositoryMock{
		CheckCredentialsImpl: func(email, password string) (bool, error) {
			return true, nil
		},
		GetUserIdByEmailImpl: func(email string) (*authRepository.UserId, error) {
			return nil, errors.New("some error")
		},
	}
	pushTokensRepositoryMock := pushNotifications_mock.RepositoryMock{}
	usersRepositoryMock := users_mock.RepositoryMock{}
	jwtServiceMock := jwt_mock.ServiceMock{}
	controller := auth.DefaultController(
		&authRepositoryMock,
		&pushTokensRepositoryMock,
		&usersRepositoryMock,
		&jwtServiceMock,
		&formatValidatorMock,
		logging.TestService(),
	)
	_, err := controller.Login(uuid.New().String(), uuid.New().String())
	if err == nil {
		t.Fatalf("err should not be nil")
	}
	if err.Code != auth.LoginErrorInternal {
		t.Fatalf("err code should be `internal`, found %v", err)
	}
}

func TestLoginGetUserNotFound(t *testing.T) {
	formatValidatorMock := formatValidation_mock.ServiceMock{}
	authRepositoryMock := auth_mock.RepositoryMock{
		CheckCredentialsImpl: func(email, password string) (bool, error) {
			return true, nil
		},
		GetUserIdByEmailImpl: func(email string) (*authRepository.UserId, error) {
			return nil, nil
		},
	}
	pushTokensRepositoryMock := pushNotifications_mock.RepositoryMock{}
	usersRepositoryMock := users_mock.RepositoryMock{}
	jwtServiceMock := jwt_mock.ServiceMock{}
	controller := auth.DefaultController(
		&authRepositoryMock,
		&pushTokensRepositoryMock,
		&usersRepositoryMock,
		&jwtServiceMock,
		&formatValidatorMock,
		logging.TestService(),
	)
	_, err := controller.Login(uuid.New().String(), uuid.New().String())
	if err == nil {
		t.Fatalf("err should not be nil")
	}
	if err.Code != auth.LoginErrorInternal {
		t.Fatalf("err code should be `internal`, found %v", err)
	}
}

func TestLoginIssueAccessTokenFailed(t *testing.T) {
	formatValidatorMock := formatValidation_mock.ServiceMock{}
	authRepositoryMock := auth_mock.RepositoryMock{
		CheckCredentialsImpl: func(email, password string) (bool, error) {
			return true, nil
		},
		GetUserIdByEmailImpl: func(email string) (*authRepository.UserId, error) {
			uid := uuid.New().String()
			return (*authRepository.UserId)(&uid), nil
		},
	}
	pushTokensRepositoryMock := pushNotifications_mock.RepositoryMock{}
	usersRepositoryMock := users_mock.RepositoryMock{}
	jwtServiceMock := jwt_mock.ServiceMock{
		IssueAccessTokenImpl: func(subject jwt.Subject) (jwt.AccessToken, *jwt.Error) {
			return jwt.AccessToken(uuid.New().String()), &jwt.Error{}
		},
		IssueRefreshTokenImpl: func(subject jwt.Subject) (jwt.RefreshToken, *jwt.Error) {
			return jwt.RefreshToken(uuid.New().String()), nil
		},
	}
	controller := auth.DefaultController(
		&authRepositoryMock,
		&pushTokensRepositoryMock,
		&usersRepositoryMock,
		&jwtServiceMock,
		&formatValidatorMock,
		logging.TestService(),
	)
	_, err := controller.Login(uuid.New().String(), uuid.New().String())
	if err == nil {
		t.Fatalf("err should not be nil")
	}
	if err.Code != auth.LoginErrorInternal {
		t.Fatalf("err code should be `internal`, found %v", err)
	}
}

func TestLoginIssueRefreshTokenFailed(t *testing.T) {
	formatValidatorMock := formatValidation_mock.ServiceMock{}
	authRepositoryMock := auth_mock.RepositoryMock{
		CheckCredentialsImpl: func(email, password string) (bool, error) {
			return true, nil
		},
		GetUserIdByEmailImpl: func(email string) (*authRepository.UserId, error) {
			uid := uuid.New().String()
			return (*authRepository.UserId)(&uid), nil
		},
	}
	pushTokensRepositoryMock := pushNotifications_mock.RepositoryMock{}
	usersRepositoryMock := users_mock.RepositoryMock{}
	jwtServiceMock := jwt_mock.ServiceMock{
		IssueAccessTokenImpl: func(subject jwt.Subject) (jwt.AccessToken, *jwt.Error) {
			return jwt.AccessToken(uuid.New().String()), nil
		},
		IssueRefreshTokenImpl: func(subject jwt.Subject) (jwt.RefreshToken, *jwt.Error) {
			return jwt.RefreshToken(uuid.New().String()), &jwt.Error{}
		},
	}
	controller := auth.DefaultController(
		&authRepositoryMock,
		&pushTokensRepositoryMock,
		&usersRepositoryMock,
		&jwtServiceMock,
		&formatValidatorMock,
		logging.TestService(),
	)
	_, err := controller.Login(uuid.New().String(), uuid.New().String())
	if err == nil {
		t.Fatalf("err should not be nil")
	}
	if err.Code != auth.LoginErrorInternal {
		t.Fatalf("err code should be `internal`, found %v", err)
	}
}

func TestLoginUpdateTokenFailed(t *testing.T) {
	formatValidatorMock := formatValidation_mock.ServiceMock{}
	authRepositoryMock := auth_mock.RepositoryMock{
		CheckCredentialsImpl: func(email, password string) (bool, error) {
			return true, nil
		},
		GetUserIdByEmailImpl: func(email string) (*authRepository.UserId, error) {
			uid := uuid.New().String()
			return (*authRepository.UserId)(&uid), nil
		},
		UpdateRefreshTokenImpl: func(uid authRepository.UserId, token string) repositories.MutationWorkItem {
			return repositories.MutationWorkItem{
				Perform: func() error {
					return errors.New("some error")
				},
			}
		},
	}
	pushTokensRepositoryMock := pushNotifications_mock.RepositoryMock{}
	usersRepositoryMock := users_mock.RepositoryMock{}
	jwtServiceMock := jwt_mock.ServiceMock{
		IssueAccessTokenImpl: func(subject jwt.Subject) (jwt.AccessToken, *jwt.Error) {
			return jwt.AccessToken(uuid.New().String()), nil
		},
		IssueRefreshTokenImpl: func(subject jwt.Subject) (jwt.RefreshToken, *jwt.Error) {
			return jwt.RefreshToken(uuid.New().String()), nil
		},
	}
	controller := auth.DefaultController(
		&authRepositoryMock,
		&pushTokensRepositoryMock,
		&usersRepositoryMock,
		&jwtServiceMock,
		&formatValidatorMock,
		logging.TestService(),
	)
	_, err := controller.Login(uuid.New().String(), uuid.New().String())
	if err == nil {
		t.Fatalf("err should not be nil")
	}
	if err.Code != auth.LoginErrorInternal {
		t.Fatalf("err code should be `internal`, found %v", err)
	}
}

func TestLoginOk(t *testing.T) {
	updateTokenCalls := 0
	formatValidatorMock := formatValidation_mock.ServiceMock{}
	authRepositoryMock := auth_mock.RepositoryMock{
		CheckCredentialsImpl: func(email, password string) (bool, error) {
			return true, nil
		},
		GetUserIdByEmailImpl: func(email string) (*authRepository.UserId, error) {
			uid := uuid.New().String()
			return (*authRepository.UserId)(&uid), nil
		},
		UpdateRefreshTokenImpl: func(uid authRepository.UserId, token string) repositories.MutationWorkItem {
			return repositories.MutationWorkItem{
				Perform: func() error {
					updateTokenCalls += 1
					return nil
				},
			}
		},
	}
	pushTokensRepositoryMock := pushNotifications_mock.RepositoryMock{}
	usersRepositoryMock := users_mock.RepositoryMock{}
	jwtServiceMock := jwt_mock.ServiceMock{
		IssueAccessTokenImpl: func(subject jwt.Subject) (jwt.AccessToken, *jwt.Error) {
			return jwt.AccessToken(uuid.New().String()), nil
		},
		IssueRefreshTokenImpl: func(subject jwt.Subject) (jwt.RefreshToken, *jwt.Error) {
			return jwt.RefreshToken(uuid.New().String()), nil
		},
	}
	controller := auth.DefaultController(
		&authRepositoryMock,
		&pushTokensRepositoryMock,
		&usersRepositoryMock,
		&jwtServiceMock,
		&formatValidatorMock,
		logging.TestService(),
	)
	_, err := controller.Login(uuid.New().String(), uuid.New().String())
	if err != nil {
		t.Fatalf("err should be nil, found %v", err)
	}
	if updateTokenCalls != 1 {
		t.Fatalf("should update token once")
	}
}

func TestRefreshTokenExpired(t *testing.T) {
	formatValidatorMock := formatValidation_mock.ServiceMock{}
	authRepositoryMock := auth_mock.RepositoryMock{}
	pushTokensRepositoryMock := pushNotifications_mock.RepositoryMock{}
	usersRepositoryMock := users_mock.RepositoryMock{}
	jwtServiceMock := jwt_mock.ServiceMock{
		ValidateRefreshTokenImpl: func(token jwt.RefreshToken) *jwt.Error {
			return &jwt.Error{Code: jwt.CodeTokenExpired}
		},
	}
	controller := auth.DefaultController(
		&authRepositoryMock,
		&pushTokensRepositoryMock,
		&usersRepositoryMock,
		&jwtServiceMock,
		&formatValidatorMock,
		logging.TestService(),
	)
	_, err := controller.Refresh(uuid.New().String())
	if err == nil {
		t.Fatalf("err should not be nil")
	}
	if err.Code != auth.RefreshErrorTokenExpired {
		t.Fatalf("err code should be `token expired`, found %v", err)
	}
}

func TestRefreshTokenWrong(t *testing.T) {
	formatValidatorMock := formatValidation_mock.ServiceMock{}
	authRepositoryMock := auth_mock.RepositoryMock{}
	pushTokensRepositoryMock := pushNotifications_mock.RepositoryMock{}
	usersRepositoryMock := users_mock.RepositoryMock{}
	jwtServiceMock := jwt_mock.ServiceMock{
		ValidateRefreshTokenImpl: func(token jwt.RefreshToken) *jwt.Error {
			return &jwt.Error{Code: jwt.CodeTokenInvalid}
		},
	}
	controller := auth.DefaultController(
		&authRepositoryMock,
		&pushTokensRepositoryMock,
		&usersRepositoryMock,
		&jwtServiceMock,
		&formatValidatorMock,
		logging.TestService(),
	)
	_, err := controller.Refresh(uuid.New().String())
	if err == nil {
		t.Fatalf("err should not be nil")
	}
	if err.Code != auth.RefreshErrorTokenIsWrong {
		t.Fatalf("err code should be `token wrong`, found %v", err)
	}
}

func TestRefreshUnableToValidateToken(t *testing.T) {
	formatValidatorMock := formatValidation_mock.ServiceMock{}
	authRepositoryMock := auth_mock.RepositoryMock{}
	pushTokensRepositoryMock := pushNotifications_mock.RepositoryMock{}
	usersRepositoryMock := users_mock.RepositoryMock{}
	jwtServiceMock := jwt_mock.ServiceMock{
		ValidateRefreshTokenImpl: func(token jwt.RefreshToken) *jwt.Error {
			return &jwt.Error{Code: jwt.CodeInternal}
		},
	}
	controller := auth.DefaultController(
		&authRepositoryMock,
		&pushTokensRepositoryMock,
		&usersRepositoryMock,
		&jwtServiceMock,
		&formatValidatorMock,
		logging.TestService(),
	)
	_, err := controller.Refresh(uuid.New().String())
	if err == nil {
		t.Fatalf("err should not be nil")
	}
	if err.Code != auth.RefreshErrorInternal {
		t.Fatalf("err code should be `internal`, found %v", err)
	}
}

func TestRefreshUnableToGetSubject(t *testing.T) {
	formatValidatorMock := formatValidation_mock.ServiceMock{}
	authRepositoryMock := auth_mock.RepositoryMock{}
	pushTokensRepositoryMock := pushNotifications_mock.RepositoryMock{}
	usersRepositoryMock := users_mock.RepositoryMock{}
	jwtServiceMock := jwt_mock.ServiceMock{
		ValidateRefreshTokenImpl: func(token jwt.RefreshToken) *jwt.Error {
			return nil
		},
		GetRefreshTokenSubjectImpl: func(token jwt.RefreshToken) (jwt.Subject, *jwt.Error) {
			return jwt.Subject(uuid.New().String()), &jwt.Error{}
		},
	}
	controller := auth.DefaultController(
		&authRepositoryMock,
		&pushTokensRepositoryMock,
		&usersRepositoryMock,
		&jwtServiceMock,
		&formatValidatorMock,
		logging.TestService(),
	)
	_, err := controller.Refresh(uuid.New().String())
	if err == nil {
		t.Fatalf("err should not be nil")
	}
	if err.Code != auth.RefreshErrorInternal {
		t.Fatalf("err code should be `internal`, found %v", err)
	}
}

func TestRefreshUnableToGetCurrentToken(t *testing.T) {
	formatValidatorMock := formatValidation_mock.ServiceMock{}
	authRepositoryMock := auth_mock.RepositoryMock{
		GetUserInfoImpl: func(uid authRepository.UserId) (authRepository.UserInfo, error) {
			return authRepository.UserInfo{}, errors.New("some error")
		},
	}
	pushTokensRepositoryMock := pushNotifications_mock.RepositoryMock{}
	usersRepositoryMock := users_mock.RepositoryMock{}
	jwtServiceMock := jwt_mock.ServiceMock{
		ValidateRefreshTokenImpl: func(token jwt.RefreshToken) *jwt.Error {
			return nil
		},
		GetRefreshTokenSubjectImpl: func(token jwt.RefreshToken) (jwt.Subject, *jwt.Error) {
			return jwt.Subject(uuid.New().String()), nil
		},
	}
	controller := auth.DefaultController(
		&authRepositoryMock,
		&pushTokensRepositoryMock,
		&usersRepositoryMock,
		&jwtServiceMock,
		&formatValidatorMock,
		logging.TestService(),
	)
	_, err := controller.Refresh(uuid.New().String())
	if err == nil {
		t.Fatalf("err should not be nil")
	}
	if err.Code != auth.RefreshErrorInternal {
		t.Fatalf("err code should be `internal`, found %v", err)
	}
}

func TestRefreshTokensDidNotMatch(t *testing.T) {
	currentToken := uuid.New().String()
	formatValidatorMock := formatValidation_mock.ServiceMock{}
	authRepositoryMock := auth_mock.RepositoryMock{
		GetUserInfoImpl: func(uid authRepository.UserId) (authRepository.UserInfo, error) {
			return authRepository.UserInfo{
				RefreshToken: currentToken,
			}, nil
		},
	}
	pushTokensRepositoryMock := pushNotifications_mock.RepositoryMock{}
	usersRepositoryMock := users_mock.RepositoryMock{}
	jwtServiceMock := jwt_mock.ServiceMock{
		ValidateRefreshTokenImpl: func(token jwt.RefreshToken) *jwt.Error {
			return nil
		},
		GetRefreshTokenSubjectImpl: func(token jwt.RefreshToken) (jwt.Subject, *jwt.Error) {
			return jwt.Subject(uuid.New().String()), nil
		},
	}
	controller := auth.DefaultController(
		&authRepositoryMock,
		&pushTokensRepositoryMock,
		&usersRepositoryMock,
		&jwtServiceMock,
		&formatValidatorMock,
		logging.TestService(),
	)
	_, err := controller.Refresh(uuid.New().String())
	if err == nil {
		t.Fatalf("err should not be nil")
	}
	if err.Code != auth.RefreshErrorTokenIsWrong {
		t.Fatalf("err code should be `token wrong`, found %v", err)
	}
}

func TestRefreshIssueAccessTokenFailed(t *testing.T) {
	currentToken := uuid.New().String()
	formatValidatorMock := formatValidation_mock.ServiceMock{}
	authRepositoryMock := auth_mock.RepositoryMock{
		GetUserInfoImpl: func(uid authRepository.UserId) (authRepository.UserInfo, error) {
			return authRepository.UserInfo{
				RefreshToken: currentToken,
			}, nil
		},
	}
	pushTokensRepositoryMock := pushNotifications_mock.RepositoryMock{}
	usersRepositoryMock := users_mock.RepositoryMock{}
	jwtServiceMock := jwt_mock.ServiceMock{
		ValidateRefreshTokenImpl: func(token jwt.RefreshToken) *jwt.Error {
			return nil
		},
		GetRefreshTokenSubjectImpl: func(token jwt.RefreshToken) (jwt.Subject, *jwt.Error) {
			return jwt.Subject(uuid.New().String()), nil
		},
		IssueAccessTokenImpl: func(subject jwt.Subject) (jwt.AccessToken, *jwt.Error) {
			return jwt.AccessToken(uuid.New().String()), &jwt.Error{}
		},
		IssueRefreshTokenImpl: func(subject jwt.Subject) (jwt.RefreshToken, *jwt.Error) {
			return jwt.RefreshToken(uuid.New().String()), nil
		},
	}
	controller := auth.DefaultController(
		&authRepositoryMock,
		&pushTokensRepositoryMock,
		&usersRepositoryMock,
		&jwtServiceMock,
		&formatValidatorMock,
		logging.TestService(),
	)
	_, err := controller.Refresh(currentToken)
	if err == nil {
		t.Fatalf("err should not be nil")
	}
	if err.Code != auth.RefreshErrorInternal {
		t.Fatalf("err code should be `internal`, found %v", err)
	}
}

func TestRefreshIssueRefreshTokenFailed(t *testing.T) {
	currentToken := uuid.New().String()
	formatValidatorMock := formatValidation_mock.ServiceMock{}
	authRepositoryMock := auth_mock.RepositoryMock{
		GetUserInfoImpl: func(uid authRepository.UserId) (authRepository.UserInfo, error) {
			return authRepository.UserInfo{
				RefreshToken: currentToken,
			}, nil
		},
	}
	pushTokensRepositoryMock := pushNotifications_mock.RepositoryMock{}
	usersRepositoryMock := users_mock.RepositoryMock{}
	jwtServiceMock := jwt_mock.ServiceMock{
		ValidateRefreshTokenImpl: func(token jwt.RefreshToken) *jwt.Error {
			return nil
		},
		GetRefreshTokenSubjectImpl: func(token jwt.RefreshToken) (jwt.Subject, *jwt.Error) {
			return jwt.Subject(uuid.New().String()), nil
		},
		IssueAccessTokenImpl: func(subject jwt.Subject) (jwt.AccessToken, *jwt.Error) {
			return jwt.AccessToken(uuid.New().String()), nil
		},
		IssueRefreshTokenImpl: func(subject jwt.Subject) (jwt.RefreshToken, *jwt.Error) {
			return jwt.RefreshToken(uuid.New().String()), &jwt.Error{}
		},
	}
	controller := auth.DefaultController(
		&authRepositoryMock,
		&pushTokensRepositoryMock,
		&usersRepositoryMock,
		&jwtServiceMock,
		&formatValidatorMock,
		logging.TestService(),
	)
	_, err := controller.Refresh(currentToken)
	if err == nil {
		t.Fatalf("err should not be nil")
	}
	if err.Code != auth.RefreshErrorInternal {
		t.Fatalf("err code should be `internal`, found %v", err)
	}
}

func TestRefreshUpdateRefreshTokenFailed(t *testing.T) {
	currentToken := uuid.New().String()
	formatValidatorMock := formatValidation_mock.ServiceMock{}
	authRepositoryMock := auth_mock.RepositoryMock{
		GetUserInfoImpl: func(uid authRepository.UserId) (authRepository.UserInfo, error) {
			return authRepository.UserInfo{
				RefreshToken: currentToken,
			}, nil
		},
		UpdateRefreshTokenImpl: func(uid authRepository.UserId, token string) repositories.MutationWorkItem {
			return repositories.MutationWorkItem{
				Perform: func() error {
					return errors.New("some error")
				},
			}
		},
	}
	pushTokensRepositoryMock := pushNotifications_mock.RepositoryMock{}
	usersRepositoryMock := users_mock.RepositoryMock{}
	jwtServiceMock := jwt_mock.ServiceMock{
		ValidateRefreshTokenImpl: func(token jwt.RefreshToken) *jwt.Error {
			return nil
		},
		GetRefreshTokenSubjectImpl: func(token jwt.RefreshToken) (jwt.Subject, *jwt.Error) {
			return jwt.Subject(uuid.New().String()), nil
		},
		IssueAccessTokenImpl: func(subject jwt.Subject) (jwt.AccessToken, *jwt.Error) {
			return jwt.AccessToken(uuid.New().String()), nil
		},
		IssueRefreshTokenImpl: func(subject jwt.Subject) (jwt.RefreshToken, *jwt.Error) {
			return jwt.RefreshToken(uuid.New().String()), nil
		},
	}
	controller := auth.DefaultController(
		&authRepositoryMock,
		&pushTokensRepositoryMock,
		&usersRepositoryMock,
		&jwtServiceMock,
		&formatValidatorMock,
		logging.TestService(),
	)
	_, err := controller.Refresh(currentToken)
	if err == nil {
		t.Fatalf("err should not be nil")
	}
	if err.Code != auth.RefreshErrorInternal {
		t.Fatalf("err code should be `internal`, found %v", err)
	}
}

func TestRefreshOk(t *testing.T) {
	updateRefreshTokenCalls := 0
	currentToken := uuid.New().String()
	formatValidatorMock := formatValidation_mock.ServiceMock{}
	authRepositoryMock := auth_mock.RepositoryMock{
		GetUserInfoImpl: func(uid authRepository.UserId) (authRepository.UserInfo, error) {
			return authRepository.UserInfo{
				RefreshToken: currentToken,
			}, nil
		},
		UpdateRefreshTokenImpl: func(uid authRepository.UserId, token string) repositories.MutationWorkItem {
			return repositories.MutationWorkItem{
				Perform: func() error {
					updateRefreshTokenCalls += 1
					return nil
				},
			}
		},
	}
	pushTokensRepositoryMock := pushNotifications_mock.RepositoryMock{}
	usersRepositoryMock := users_mock.RepositoryMock{}
	jwtServiceMock := jwt_mock.ServiceMock{
		ValidateRefreshTokenImpl: func(token jwt.RefreshToken) *jwt.Error {
			return nil
		},
		GetRefreshTokenSubjectImpl: func(token jwt.RefreshToken) (jwt.Subject, *jwt.Error) {
			return jwt.Subject(uuid.New().String()), nil
		},
		IssueAccessTokenImpl: func(subject jwt.Subject) (jwt.AccessToken, *jwt.Error) {
			return jwt.AccessToken(uuid.New().String()), nil
		},
		IssueRefreshTokenImpl: func(subject jwt.Subject) (jwt.RefreshToken, *jwt.Error) {
			return jwt.RefreshToken(uuid.New().String()), nil
		},
	}
	controller := auth.DefaultController(
		&authRepositoryMock,
		&pushTokensRepositoryMock,
		&usersRepositoryMock,
		&jwtServiceMock,
		&formatValidatorMock,
		logging.TestService(),
	)
	_, err := controller.Refresh(currentToken)
	if err != nil {
		t.Fatalf("err should be nil, found %v", err)
	}
	if updateRefreshTokenCalls != 1 {
		t.Fatalf("token should be updated once, found %d", updateRefreshTokenCalls)
	}
}

func TestLogoutIssueNewRefreshTokenFailed(t *testing.T) {
	formatValidatorMock := formatValidation_mock.ServiceMock{}
	authRepositoryMock := auth_mock.RepositoryMock{}
	pushTokensRepositoryMock := pushNotifications_mock.RepositoryMock{}
	usersRepositoryMock := users_mock.RepositoryMock{}
	jwtServiceMock := jwt_mock.ServiceMock{
		IssueRefreshTokenImpl: func(subject jwt.Subject) (jwt.RefreshToken, *jwt.Error) {
			return jwt.RefreshToken(uuid.New().String()), &jwt.Error{}
		},
	}
	controller := auth.DefaultController(
		&authRepositoryMock,
		&pushTokensRepositoryMock,
		&usersRepositoryMock,
		&jwtServiceMock,
		&formatValidatorMock,
		logging.TestService(),
	)
	err := controller.Logout(auth.UserId(uuid.New().String()))
	if err == nil {
		t.Fatalf("err should not be nil")
	}
	if err.Code != auth.LogoutErrorInternal {
		t.Fatalf("err code should be `internal`, found %v", err)
	}
}

func TestLogoutUpdateRefreshTokenFailed(t *testing.T) {
	formatValidatorMock := formatValidation_mock.ServiceMock{}
	authRepositoryMock := auth_mock.RepositoryMock{
		UpdateRefreshTokenImpl: func(uid authRepository.UserId, token string) repositories.MutationWorkItem {
			return repositories.MutationWorkItem{
				Perform: func() error {
					return errors.New("some error")
				},
			}
		},
	}
	pushTokensRepositoryMock := pushNotifications_mock.RepositoryMock{}
	usersRepositoryMock := users_mock.RepositoryMock{}
	jwtServiceMock := jwt_mock.ServiceMock{
		IssueRefreshTokenImpl: func(subject jwt.Subject) (jwt.RefreshToken, *jwt.Error) {
			return jwt.RefreshToken(uuid.New().String()), nil
		},
	}
	controller := auth.DefaultController(
		&authRepositoryMock,
		&pushTokensRepositoryMock,
		&usersRepositoryMock,
		&jwtServiceMock,
		&formatValidatorMock,
		logging.TestService(),
	)
	err := controller.Logout(auth.UserId(uuid.New().String()))
	if err == nil {
		t.Fatalf("err should not be nil")
	}
	if err.Code != auth.LogoutErrorInternal {
		t.Fatalf("err code should be `internal`, found %v", err)
	}
}

func TestLogoutOk(t *testing.T) {
	updateRefreshTokenCalls := 0
	formatValidatorMock := formatValidation_mock.ServiceMock{}
	authRepositoryMock := auth_mock.RepositoryMock{
		UpdateRefreshTokenImpl: func(uid authRepository.UserId, token string) repositories.MutationWorkItem {
			return repositories.MutationWorkItem{
				Perform: func() error {
					updateRefreshTokenCalls += 1
					return nil
				},
			}
		},
	}
	pushTokensRepositoryMock := pushNotifications_mock.RepositoryMock{}
	usersRepositoryMock := users_mock.RepositoryMock{}
	jwtServiceMock := jwt_mock.ServiceMock{
		IssueRefreshTokenImpl: func(subject jwt.Subject) (jwt.RefreshToken, *jwt.Error) {
			return jwt.RefreshToken(uuid.New().String()), nil
		},
	}
	controller := auth.DefaultController(
		&authRepositoryMock,
		&pushTokensRepositoryMock,
		&usersRepositoryMock,
		&jwtServiceMock,
		&formatValidatorMock,
		logging.TestService(),
	)
	err := controller.Logout(auth.UserId(uuid.New().String()))
	if err != nil {
		t.Fatalf("err should be nil, found %v", err)
	}
	if updateRefreshTokenCalls != 1 {
		t.Fatalf("token should be updated once, found %d", updateRefreshTokenCalls)
	}
}

func TestUpdateEmailWrongFormat(t *testing.T) {
	formatValidatorMock := formatValidation_mock.ServiceMock{
		ValidateEmailFormatImpl: func(email string) error {
			return errors.New("some error")
		},
	}
	authRepositoryMock := auth_mock.RepositoryMock{}
	pushTokensRepositoryMock := pushNotifications_mock.RepositoryMock{}
	usersRepositoryMock := users_mock.RepositoryMock{}
	jwtServiceMock := jwt_mock.ServiceMock{}
	controller := auth.DefaultController(
		&authRepositoryMock,
		&pushTokensRepositoryMock,
		&usersRepositoryMock,
		&jwtServiceMock,
		&formatValidatorMock,
		logging.TestService(),
	)
	_, err := controller.UpdateEmail(uuid.New().String(), auth.UserId(uuid.New().String()))
	if err == nil {
		t.Fatalf("err should not be nil")
	}
	if err.Code != auth.UpdateEmailErrorWrongFormat {
		t.Fatalf("err code should be `wrong format`, found %v", err)
	}
}

func TestUpdateEmailGetUserFailed(t *testing.T) {
	formatValidatorMock := formatValidation_mock.ServiceMock{
		ValidateEmailFormatImpl: func(email string) error {
			return nil
		},
	}
	authRepositoryMock := auth_mock.RepositoryMock{
		GetUserIdByEmailImpl: func(email string) (*authRepository.UserId, error) {
			return nil, errors.New("some error")
		},
	}
	pushTokensRepositoryMock := pushNotifications_mock.RepositoryMock{}
	usersRepositoryMock := users_mock.RepositoryMock{}
	jwtServiceMock := jwt_mock.ServiceMock{}
	controller := auth.DefaultController(
		&authRepositoryMock,
		&pushTokensRepositoryMock,
		&usersRepositoryMock,
		&jwtServiceMock,
		&formatValidatorMock,
		logging.TestService(),
	)
	_, err := controller.UpdateEmail(uuid.New().String(), auth.UserId(uuid.New().String()))
	if err == nil {
		t.Fatalf("err should not be nil")
	}
	if err.Code != auth.UpdateEmailErrorInternal {
		t.Fatalf("err code should be `internal`, found %v", err)
	}
}

func TestUpdateEmailAlreadyTaken(t *testing.T) {
	formatValidatorMock := formatValidation_mock.ServiceMock{
		ValidateEmailFormatImpl: func(email string) error {
			return nil
		},
	}
	authRepositoryMock := auth_mock.RepositoryMock{
		GetUserIdByEmailImpl: func(email string) (*authRepository.UserId, error) {
			id := uuid.New().String()
			return (*authRepository.UserId)(&id), nil
		},
	}
	pushTokensRepositoryMock := pushNotifications_mock.RepositoryMock{}
	usersRepositoryMock := users_mock.RepositoryMock{}
	jwtServiceMock := jwt_mock.ServiceMock{}
	controller := auth.DefaultController(
		&authRepositoryMock,
		&pushTokensRepositoryMock,
		&usersRepositoryMock,
		&jwtServiceMock,
		&formatValidatorMock,
		logging.TestService(),
	)
	_, err := controller.UpdateEmail(uuid.New().String(), auth.UserId(uuid.New().String()))
	if err == nil {
		t.Fatalf("err should not be nil")
	}
	if err.Code != auth.UpdateEmailErrorAlreadyTaken {
		t.Fatalf("err code should be `already taken`, found %v", err)
	}
}

func TestUpdateEmailIssueAccessTokenFailed(t *testing.T) {
	formatValidatorMock := formatValidation_mock.ServiceMock{
		ValidateEmailFormatImpl: func(email string) error {
			return nil
		},
	}
	authRepositoryMock := auth_mock.RepositoryMock{
		GetUserIdByEmailImpl: func(email string) (*authRepository.UserId, error) {
			return nil, nil
		},
	}
	pushTokensRepositoryMock := pushNotifications_mock.RepositoryMock{}
	usersRepositoryMock := users_mock.RepositoryMock{}
	jwtServiceMock := jwt_mock.ServiceMock{
		IssueAccessTokenImpl: func(subject jwt.Subject) (jwt.AccessToken, *jwt.Error) {
			return jwt.AccessToken(uuid.New().String()), &jwt.Error{}
		},
		IssueRefreshTokenImpl: func(subject jwt.Subject) (jwt.RefreshToken, *jwt.Error) {
			return jwt.RefreshToken(uuid.New().String()), nil
		},
	}
	controller := auth.DefaultController(
		&authRepositoryMock,
		&pushTokensRepositoryMock,
		&usersRepositoryMock,
		&jwtServiceMock,
		&formatValidatorMock,
		logging.TestService(),
	)
	_, err := controller.UpdateEmail(uuid.New().String(), auth.UserId(uuid.New().String()))
	if err == nil {
		t.Fatalf("err should not be nil")
	}
	if err.Code != auth.UpdateEmailErrorInternal {
		t.Fatalf("err code should be `internal`, found %v", err)
	}
}

func TestUpdateEmailIssueRefreshTokenFailed(t *testing.T) {
	formatValidatorMock := formatValidation_mock.ServiceMock{
		ValidateEmailFormatImpl: func(email string) error {
			return nil
		},
	}
	authRepositoryMock := auth_mock.RepositoryMock{
		GetUserIdByEmailImpl: func(email string) (*authRepository.UserId, error) {
			return nil, nil
		},
	}
	pushTokensRepositoryMock := pushNotifications_mock.RepositoryMock{}
	usersRepositoryMock := users_mock.RepositoryMock{}
	jwtServiceMock := jwt_mock.ServiceMock{
		IssueAccessTokenImpl: func(subject jwt.Subject) (jwt.AccessToken, *jwt.Error) {
			return jwt.AccessToken(uuid.New().String()), nil
		},
		IssueRefreshTokenImpl: func(subject jwt.Subject) (jwt.RefreshToken, *jwt.Error) {
			return jwt.RefreshToken(uuid.New().String()), &jwt.Error{}
		},
	}
	controller := auth.DefaultController(
		&authRepositoryMock,
		&pushTokensRepositoryMock,
		&usersRepositoryMock,
		&jwtServiceMock,
		&formatValidatorMock,
		logging.TestService(),
	)
	_, err := controller.UpdateEmail(uuid.New().String(), auth.UserId(uuid.New().String()))
	if err == nil {
		t.Fatalf("err should not be nil")
	}
	if err.Code != auth.UpdateEmailErrorInternal {
		t.Fatalf("err code should be `internal`, found %v", err)
	}
}

func TestUpdateEmailEmailUpdateFailed(t *testing.T) {
	formatValidatorMock := formatValidation_mock.ServiceMock{
		ValidateEmailFormatImpl: func(email string) error {
			return nil
		},
	}
	authRepositoryMock := auth_mock.RepositoryMock{
		GetUserIdByEmailImpl: func(email string) (*authRepository.UserId, error) {
			return nil, nil
		},
		UpdateEmailImpl: func(uid authRepository.UserId, newEmail string) repositories.MutationWorkItem {
			return repositories.MutationWorkItem{
				Perform: func() error {
					return errors.New("some error")
				},
			}
		},
	}
	pushTokensRepositoryMock := pushNotifications_mock.RepositoryMock{}
	usersRepositoryMock := users_mock.RepositoryMock{}
	jwtServiceMock := jwt_mock.ServiceMock{
		IssueAccessTokenImpl: func(subject jwt.Subject) (jwt.AccessToken, *jwt.Error) {
			return jwt.AccessToken(uuid.New().String()), nil
		},
		IssueRefreshTokenImpl: func(subject jwt.Subject) (jwt.RefreshToken, *jwt.Error) {
			return jwt.RefreshToken(uuid.New().String()), nil
		},
	}
	controller := auth.DefaultController(
		&authRepositoryMock,
		&pushTokensRepositoryMock,
		&usersRepositoryMock,
		&jwtServiceMock,
		&formatValidatorMock,
		logging.TestService(),
	)
	_, err := controller.UpdateEmail(uuid.New().String(), auth.UserId(uuid.New().String()))
	if err == nil {
		t.Fatalf("err should not be nil")
	}
	if err.Code != auth.UpdateEmailErrorInternal {
		t.Fatalf("err code should be `internal`, found %v", err)
	}
}

func TestUpdateEmailTokenUpdateFailed(t *testing.T) {
	formatValidatorMock := formatValidation_mock.ServiceMock{
		ValidateEmailFormatImpl: func(email string) error {
			return nil
		},
	}
	updateEmailCalls := 0
	updateEmailRollbacks := 0
	authRepositoryMock := auth_mock.RepositoryMock{
		GetUserIdByEmailImpl: func(email string) (*authRepository.UserId, error) {
			return nil, nil
		},
		UpdateEmailImpl: func(uid authRepository.UserId, newEmail string) repositories.MutationWorkItem {
			return repositories.MutationWorkItem{
				Perform: func() error {
					updateEmailCalls += 1
					return nil
				},
				Rollback: func() error {
					updateEmailRollbacks += 1
					return nil
				},
			}
		},
		UpdateRefreshTokenImpl: func(uid authRepository.UserId, token string) repositories.MutationWorkItem {
			return repositories.MutationWorkItem{
				Perform: func() error {
					return errors.New("some error")
				},
			}
		},
	}
	pushTokensRepositoryMock := pushNotifications_mock.RepositoryMock{}
	usersRepositoryMock := users_mock.RepositoryMock{}
	jwtServiceMock := jwt_mock.ServiceMock{
		IssueAccessTokenImpl: func(subject jwt.Subject) (jwt.AccessToken, *jwt.Error) {
			return jwt.AccessToken(uuid.New().String()), nil
		},
		IssueRefreshTokenImpl: func(subject jwt.Subject) (jwt.RefreshToken, *jwt.Error) {
			return jwt.RefreshToken(uuid.New().String()), nil
		},
	}
	controller := auth.DefaultController(
		&authRepositoryMock,
		&pushTokensRepositoryMock,
		&usersRepositoryMock,
		&jwtServiceMock,
		&formatValidatorMock,
		logging.TestService(),
	)
	_, err := controller.UpdateEmail(uuid.New().String(), auth.UserId(uuid.New().String()))
	if err == nil {
		t.Fatalf("err should not be nil")
	}
	if err.Code != auth.UpdateEmailErrorInternal {
		t.Fatalf("err code should be `internal`, found %v", err)
	}
	if updateEmailCalls != 1 {
		t.Fatalf("should update email once, found %d", updateEmailCalls)
	}
	if updateEmailRollbacks != 1 {
		t.Fatalf("should update email rollback once, found %d", updateEmailRollbacks)
	}
}

func TestUpdateEmailOk(t *testing.T) {
	formatValidatorMock := formatValidation_mock.ServiceMock{
		ValidateEmailFormatImpl: func(email string) error {
			return nil
		},
	}
	updateEmailCalls := 0
	updateTokenCalls := 0
	authRepositoryMock := auth_mock.RepositoryMock{
		GetUserIdByEmailImpl: func(email string) (*authRepository.UserId, error) {
			return nil, nil
		},
		UpdateEmailImpl: func(uid authRepository.UserId, newEmail string) repositories.MutationWorkItem {
			return repositories.MutationWorkItem{
				Perform: func() error {
					updateEmailCalls += 1
					return nil
				},
			}
		},
		UpdateRefreshTokenImpl: func(uid authRepository.UserId, token string) repositories.MutationWorkItem {
			return repositories.MutationWorkItem{
				Perform: func() error {
					updateTokenCalls += 1
					return nil
				},
			}
		},
	}
	pushTokensRepositoryMock := pushNotifications_mock.RepositoryMock{}
	usersRepositoryMock := users_mock.RepositoryMock{}
	jwtServiceMock := jwt_mock.ServiceMock{
		IssueAccessTokenImpl: func(subject jwt.Subject) (jwt.AccessToken, *jwt.Error) {
			return jwt.AccessToken(uuid.New().String()), nil
		},
		IssueRefreshTokenImpl: func(subject jwt.Subject) (jwt.RefreshToken, *jwt.Error) {
			return jwt.RefreshToken(uuid.New().String()), nil
		},
	}
	controller := auth.DefaultController(
		&authRepositoryMock,
		&pushTokensRepositoryMock,
		&usersRepositoryMock,
		&jwtServiceMock,
		&formatValidatorMock,
		logging.TestService(),
	)
	_, err := controller.UpdateEmail(uuid.New().String(), auth.UserId(uuid.New().String()))
	if err != nil {
		t.Fatalf("err should be nil, found %v", err)
	}
	if updateEmailCalls != 1 {
		t.Fatalf("should update email once, found %d", updateEmailCalls)
	}
	if updateTokenCalls != 1 {
		t.Fatalf("should update token once, found %d", updateTokenCalls)
	}
}

func TestUpdatePasswordNewPasswordHasWrongFormat(t *testing.T) {
	formatValidatorMock := formatValidation_mock.ServiceMock{
		ValidatePasswordFormatImpl: func(password string) error {
			return errors.New("some error")
		},
	}
	authRepositoryMock := auth_mock.RepositoryMock{}
	pushTokensRepositoryMock := pushNotifications_mock.RepositoryMock{}
	usersRepositoryMock := users_mock.RepositoryMock{}
	jwtServiceMock := jwt_mock.ServiceMock{}
	controller := auth.DefaultController(
		&authRepositoryMock,
		&pushTokensRepositoryMock,
		&usersRepositoryMock,
		&jwtServiceMock,
		&formatValidatorMock,
		logging.TestService(),
	)
	_, err := controller.UpdatePassword(uuid.New().String(), uuid.New().String(), auth.UserId(uuid.New().String()))
	if err == nil {
		t.Fatalf("err should not be nil")
	}
	if err.Code != auth.UpdatePasswordErrorWrongFormat {
		t.Fatalf("err code should be `wrong format`, found %v", err)
	}
}

func TestUpdatePasswordGetUserFailed(t *testing.T) {
	formatValidatorMock := formatValidation_mock.ServiceMock{
		ValidatePasswordFormatImpl: func(password string) error {
			return nil
		},
	}
	authRepositoryMock := auth_mock.RepositoryMock{
		GetUserInfoImpl: func(uid authRepository.UserId) (authRepository.UserInfo, error) {
			return authRepository.UserInfo{}, errors.New("some error")
		},
	}
	pushTokensRepositoryMock := pushNotifications_mock.RepositoryMock{}
	usersRepositoryMock := users_mock.RepositoryMock{}
	jwtServiceMock := jwt_mock.ServiceMock{}
	controller := auth.DefaultController(
		&authRepositoryMock,
		&pushTokensRepositoryMock,
		&usersRepositoryMock,
		&jwtServiceMock,
		&formatValidatorMock,
		logging.TestService(),
	)
	_, err := controller.UpdatePassword(uuid.New().String(), uuid.New().String(), auth.UserId(uuid.New().String()))
	if err == nil {
		t.Fatalf("err should not be nil")
	}
	if err.Code != auth.UpdatePasswordErrorInternal {
		t.Fatalf("err code should be `internal`, found %v", err)
	}
}

func TestUpdatePasswordCredentialsCheckFailed(t *testing.T) {
	formatValidatorMock := formatValidation_mock.ServiceMock{
		ValidatePasswordFormatImpl: func(password string) error {
			return nil
		},
	}
	authRepositoryMock := auth_mock.RepositoryMock{
		CheckCredentialsImpl: func(email, password string) (bool, error) {
			return false, errors.New("some error")
		},
		GetUserInfoImpl: func(uid authRepository.UserId) (authRepository.UserInfo, error) {
			return authRepository.UserInfo{}, nil
		},
	}
	pushTokensRepositoryMock := pushNotifications_mock.RepositoryMock{}
	usersRepositoryMock := users_mock.RepositoryMock{}
	jwtServiceMock := jwt_mock.ServiceMock{}
	controller := auth.DefaultController(
		&authRepositoryMock,
		&pushTokensRepositoryMock,
		&usersRepositoryMock,
		&jwtServiceMock,
		&formatValidatorMock,
		logging.TestService(),
	)
	_, err := controller.UpdatePassword(uuid.New().String(), uuid.New().String(), auth.UserId(uuid.New().String()))
	if err == nil {
		t.Fatalf("err should not be nil")
	}
	if err.Code != auth.UpdatePasswordErrorInternal {
		t.Fatalf("err code should be `internal`, found %v", err)
	}
}

func TestUpdatePasswordOldPasswordIsWrong(t *testing.T) {
	formatValidatorMock := formatValidation_mock.ServiceMock{
		ValidatePasswordFormatImpl: func(password string) error {
			return nil
		},
	}
	authRepositoryMock := auth_mock.RepositoryMock{
		CheckCredentialsImpl: func(email, password string) (bool, error) {
			return false, nil
		},
		GetUserInfoImpl: func(uid authRepository.UserId) (authRepository.UserInfo, error) {
			return authRepository.UserInfo{}, nil
		},
	}
	pushTokensRepositoryMock := pushNotifications_mock.RepositoryMock{}
	usersRepositoryMock := users_mock.RepositoryMock{}
	jwtServiceMock := jwt_mock.ServiceMock{}
	controller := auth.DefaultController(
		&authRepositoryMock,
		&pushTokensRepositoryMock,
		&usersRepositoryMock,
		&jwtServiceMock,
		&formatValidatorMock,
		logging.TestService(),
	)
	_, err := controller.UpdatePassword(uuid.New().String(), uuid.New().String(), auth.UserId(uuid.New().String()))
	if err == nil {
		t.Fatalf("err should not be nil")
	}
	if err.Code != auth.UpdatePasswordErrorOldPasswordIsWrong {
		t.Fatalf("err code should be `old password is wrong`, found %v", err)
	}
}

func TestUpdatePasswordIssueAccessTokenFailed(t *testing.T) {
	formatValidatorMock := formatValidation_mock.ServiceMock{
		ValidatePasswordFormatImpl: func(password string) error {
			return nil
		},
	}
	authRepositoryMock := auth_mock.RepositoryMock{
		CheckCredentialsImpl: func(email, password string) (bool, error) {
			return true, nil
		},
		GetUserInfoImpl: func(uid authRepository.UserId) (authRepository.UserInfo, error) {
			return authRepository.UserInfo{}, nil
		},
	}
	pushTokensRepositoryMock := pushNotifications_mock.RepositoryMock{}
	usersRepositoryMock := users_mock.RepositoryMock{}
	jwtServiceMock := jwt_mock.ServiceMock{
		IssueAccessTokenImpl: func(subject jwt.Subject) (jwt.AccessToken, *jwt.Error) {
			return jwt.AccessToken(uuid.New().String()), &jwt.Error{}
		},
		IssueRefreshTokenImpl: func(subject jwt.Subject) (jwt.RefreshToken, *jwt.Error) {
			return jwt.RefreshToken(uuid.New().String()), nil
		},
	}
	controller := auth.DefaultController(
		&authRepositoryMock,
		&pushTokensRepositoryMock,
		&usersRepositoryMock,
		&jwtServiceMock,
		&formatValidatorMock,
		logging.TestService(),
	)
	_, err := controller.UpdatePassword(uuid.New().String(), uuid.New().String(), auth.UserId(uuid.New().String()))
	if err == nil {
		t.Fatalf("err should not be nil")
	}
	if err.Code != auth.UpdatePasswordErrorInternal {
		t.Fatalf("err code should be `internal`, found %v", err)
	}
}

func TestUpdatePasswordIssueRefreshTokenFailed(t *testing.T) {
	formatValidatorMock := formatValidation_mock.ServiceMock{
		ValidatePasswordFormatImpl: func(password string) error {
			return nil
		},
	}
	authRepositoryMock := auth_mock.RepositoryMock{
		CheckCredentialsImpl: func(email, password string) (bool, error) {
			return true, nil
		},
		GetUserInfoImpl: func(uid authRepository.UserId) (authRepository.UserInfo, error) {
			return authRepository.UserInfo{}, nil
		},
	}
	pushTokensRepositoryMock := pushNotifications_mock.RepositoryMock{}
	usersRepositoryMock := users_mock.RepositoryMock{}
	jwtServiceMock := jwt_mock.ServiceMock{
		IssueAccessTokenImpl: func(subject jwt.Subject) (jwt.AccessToken, *jwt.Error) {
			return jwt.AccessToken(uuid.New().String()), nil
		},
		IssueRefreshTokenImpl: func(subject jwt.Subject) (jwt.RefreshToken, *jwt.Error) {
			return jwt.RefreshToken(uuid.New().String()), &jwt.Error{}
		},
	}
	controller := auth.DefaultController(
		&authRepositoryMock,
		&pushTokensRepositoryMock,
		&usersRepositoryMock,
		&jwtServiceMock,
		&formatValidatorMock,
		logging.TestService(),
	)
	_, err := controller.UpdatePassword(uuid.New().String(), uuid.New().String(), auth.UserId(uuid.New().String()))
	if err == nil {
		t.Fatalf("err should not be nil")
	}
	if err.Code != auth.UpdatePasswordErrorInternal {
		t.Fatalf("err code should be `internal`, found %v", err)
	}
}

func TestUpdatePasswordPasswordUpdateFailed(t *testing.T) {
	formatValidatorMock := formatValidation_mock.ServiceMock{
		ValidatePasswordFormatImpl: func(password string) error {
			return nil
		},
	}
	authRepositoryMock := auth_mock.RepositoryMock{
		CheckCredentialsImpl: func(email, password string) (bool, error) {
			return true, nil
		},
		GetUserInfoImpl: func(uid authRepository.UserId) (authRepository.UserInfo, error) {
			return authRepository.UserInfo{}, nil
		},
		UpdatePasswordImpl: func(uid authRepository.UserId, newPassword string) repositories.MutationWorkItem {
			return repositories.MutationWorkItem{
				Perform: func() error {
					return errors.New("")
				},
			}
		},
	}
	pushTokensRepositoryMock := pushNotifications_mock.RepositoryMock{}
	usersRepositoryMock := users_mock.RepositoryMock{}
	jwtServiceMock := jwt_mock.ServiceMock{
		IssueAccessTokenImpl: func(subject jwt.Subject) (jwt.AccessToken, *jwt.Error) {
			return jwt.AccessToken(uuid.New().String()), nil
		},
		IssueRefreshTokenImpl: func(subject jwt.Subject) (jwt.RefreshToken, *jwt.Error) {
			return jwt.RefreshToken(uuid.New().String()), nil
		},
	}
	controller := auth.DefaultController(
		&authRepositoryMock,
		&pushTokensRepositoryMock,
		&usersRepositoryMock,
		&jwtServiceMock,
		&formatValidatorMock,
		logging.TestService(),
	)
	_, err := controller.UpdatePassword(uuid.New().String(), uuid.New().String(), auth.UserId(uuid.New().String()))
	if err == nil {
		t.Fatalf("err should not be nil")
	}
	if err.Code != auth.UpdatePasswordErrorInternal {
		t.Fatalf("err code should be `internal`, found %v", err)
	}
}

func TestUpdatePasswordTokenUpdateFailed(t *testing.T) {
	formatValidatorMock := formatValidation_mock.ServiceMock{
		ValidatePasswordFormatImpl: func(password string) error {
			return nil
		},
	}
	updatePasswordCalls := 0
	updatePasswordRollbacks := 0
	authRepositoryMock := auth_mock.RepositoryMock{
		CheckCredentialsImpl: func(email, password string) (bool, error) {
			return true, nil
		},
		GetUserInfoImpl: func(uid authRepository.UserId) (authRepository.UserInfo, error) {
			return authRepository.UserInfo{}, nil
		},
		UpdatePasswordImpl: func(uid authRepository.UserId, newPassword string) repositories.MutationWorkItem {
			return repositories.MutationWorkItem{
				Perform: func() error {
					updatePasswordCalls += 1
					return nil
				},
				Rollback: func() error {
					updatePasswordRollbacks += 1
					return nil
				},
			}
		},
		UpdateRefreshTokenImpl: func(uid authRepository.UserId, token string) repositories.MutationWorkItem {
			return repositories.MutationWorkItem{
				Perform: func() error {
					return errors.New("some error")
				},
			}
		},
	}
	pushTokensRepositoryMock := pushNotifications_mock.RepositoryMock{}
	usersRepositoryMock := users_mock.RepositoryMock{}
	jwtServiceMock := jwt_mock.ServiceMock{
		IssueAccessTokenImpl: func(subject jwt.Subject) (jwt.AccessToken, *jwt.Error) {
			return jwt.AccessToken(uuid.New().String()), nil
		},
		IssueRefreshTokenImpl: func(subject jwt.Subject) (jwt.RefreshToken, *jwt.Error) {
			return jwt.RefreshToken(uuid.New().String()), nil
		},
	}
	controller := auth.DefaultController(
		&authRepositoryMock,
		&pushTokensRepositoryMock,
		&usersRepositoryMock,
		&jwtServiceMock,
		&formatValidatorMock,
		logging.TestService(),
	)
	_, err := controller.UpdatePassword(uuid.New().String(), uuid.New().String(), auth.UserId(uuid.New().String()))
	if err == nil {
		t.Fatalf("err should not be nil")
	}
	if err.Code != auth.UpdatePasswordErrorInternal {
		t.Fatalf("err code should be `internal`, found %v", err)
	}
	if updatePasswordCalls != 1 {
		t.Fatalf("password should be updated once, found %d", updatePasswordCalls)
	}
	if updatePasswordRollbacks != 1 {
		t.Fatalf("password update should be rolled back once, found %d", updatePasswordRollbacks)
	}
}

func TestUpdatePasswordOk(t *testing.T) {
	formatValidatorMock := formatValidation_mock.ServiceMock{
		ValidatePasswordFormatImpl: func(password string) error {
			return nil
		},
	}
	updatePasswordCalls := 0
	updateTokenCalls := 0
	authRepositoryMock := auth_mock.RepositoryMock{
		CheckCredentialsImpl: func(email, password string) (bool, error) {
			return true, nil
		},
		GetUserInfoImpl: func(uid authRepository.UserId) (authRepository.UserInfo, error) {
			return authRepository.UserInfo{}, nil
		},
		UpdatePasswordImpl: func(uid authRepository.UserId, newPassword string) repositories.MutationWorkItem {
			return repositories.MutationWorkItem{
				Perform: func() error {
					updatePasswordCalls += 1
					return nil
				},
			}
		},
		UpdateRefreshTokenImpl: func(uid authRepository.UserId, token string) repositories.MutationWorkItem {
			return repositories.MutationWorkItem{
				Perform: func() error {
					updateTokenCalls += 1
					return nil
				},
			}
		},
	}
	pushTokensRepositoryMock := pushNotifications_mock.RepositoryMock{}
	usersRepositoryMock := users_mock.RepositoryMock{}
	jwtServiceMock := jwt_mock.ServiceMock{
		IssueAccessTokenImpl: func(subject jwt.Subject) (jwt.AccessToken, *jwt.Error) {
			return jwt.AccessToken(uuid.New().String()), nil
		},
		IssueRefreshTokenImpl: func(subject jwt.Subject) (jwt.RefreshToken, *jwt.Error) {
			return jwt.RefreshToken(uuid.New().String()), nil
		},
	}
	controller := auth.DefaultController(
		&authRepositoryMock,
		&pushTokensRepositoryMock,
		&usersRepositoryMock,
		&jwtServiceMock,
		&formatValidatorMock,
		logging.TestService(),
	)
	_, err := controller.UpdatePassword(uuid.New().String(), uuid.New().String(), auth.UserId(uuid.New().String()))
	if err != nil {
		t.Fatalf("err should be nil, found %v", err)
	}
	if updatePasswordCalls != 1 {
		t.Fatalf("password should be updated once, found %d", updatePasswordCalls)
	}
	if updateTokenCalls != 1 {
		t.Fatalf("token update should be once, found %d", updateTokenCalls)
	}
}

func TestRegisterForPushNotificationsFailedToStoreToken(t *testing.T) {
	formatValidatorMock := formatValidation_mock.ServiceMock{}
	authRepositoryMock := auth_mock.RepositoryMock{}
	pushTokensRepositoryMock := pushNotifications_mock.RepositoryMock{
		StorePushTokenImpl: func(uid pushNotifications.UserId, token string) repositories.MutationWorkItem {
			return repositories.MutationWorkItem{
				Perform: func() error {
					return errors.New("some error")
				},
			}
		},
	}
	usersRepositoryMock := users_mock.RepositoryMock{}
	jwtServiceMock := jwt_mock.ServiceMock{}
	controller := auth.DefaultController(
		&authRepositoryMock,
		&pushTokensRepositoryMock,
		&usersRepositoryMock,
		&jwtServiceMock,
		&formatValidatorMock,
		logging.TestService(),
	)
	err := controller.RegisterForPushNotifications(uuid.New().String(), auth.UserId(uuid.New().String()))
	if err == nil {
		t.Fatalf("err should not be nil")
	}
	if err.Code != auth.RegisterForPushNotificationsErrorInternal {
		t.Fatalf("err code should be `internal`, found %v", err)
	}
}

func TestRegisterForPushNotificationsOk(t *testing.T) {
	formatValidatorMock := formatValidation_mock.ServiceMock{}
	authRepositoryMock := auth_mock.RepositoryMock{}
	storeTokenCalls := 0
	pushTokensRepositoryMock := pushNotifications_mock.RepositoryMock{
		StorePushTokenImpl: func(uid pushNotifications.UserId, token string) repositories.MutationWorkItem {
			return repositories.MutationWorkItem{
				Perform: func() error {
					storeTokenCalls += 1
					return nil
				},
			}
		},
	}
	usersRepositoryMock := users_mock.RepositoryMock{}
	jwtServiceMock := jwt_mock.ServiceMock{}
	controller := auth.DefaultController(
		&authRepositoryMock,
		&pushTokensRepositoryMock,
		&usersRepositoryMock,
		&jwtServiceMock,
		&formatValidatorMock,
		logging.TestService(),
	)
	err := controller.RegisterForPushNotifications(uuid.New().String(), auth.UserId(uuid.New().String()))
	if err != nil {
		t.Fatalf("err should be nil, found %v", err)
	}
	if storeTokenCalls != 1 {
		t.Fatalf("store should be called once, found %d", storeTokenCalls)
	}
}
