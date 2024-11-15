package defaultController_test

import (
	"errors"
	"testing"
	"verni/internal/controllers/verification"
	defaultController "verni/internal/controllers/verification/default"
	"verni/internal/repositories"
	"verni/internal/repositories/auth"
	auth_mock "verni/internal/repositories/auth/mock"
	verification_mock "verni/internal/repositories/verification/mock"
	emailSender_mock "verni/internal/services/emailSender/mock"
	"verni/internal/services/logging"

	"github.com/google/uuid"
)

func TestSendConfirmationCodeNotDelivered(t *testing.T) {
	storeCalled := 0
	storeRolledBack := 0
	triedToSend := 0
	emailSenderMock := emailSender_mock.ServiceMock{
		SendImpl: func(subject, email string) error {
			triedToSend += 1
			return errors.New("some error")
		},
	}
	verificationMock := verification_mock.RepositoryMock{
		StoreEmailVerificationCodeImpl: func(email string, code string) repositories.MutationWorkItem {
			return repositories.MutationWorkItem{
				Perform: func() error {
					storeCalled += 1
					return nil
				},
				Rollback: func() error {
					storeRolledBack += 1
					return nil
				},
			}
		},
	}
	authMock := auth_mock.RepositoryMock{
		GetUserInfoImpl: func(uid auth.UserId) (auth.UserInfo, error) {
			return auth.UserInfo{}, nil
		},
	}
	controller := defaultController.New(
		&verificationMock,
		&authMock,
		&emailSenderMock,
		logging.TestService(),
	)
	err := controller.SendConfirmationCode(verification.UserId(uuid.New().String()))
	if err == nil {
		t.Fatalf("`err` should not be nil")
	}
	if err.Code != verification.SendConfirmationCodeErrorNotDelivered {
		t.Fatalf("unexpected error code, expected `not delivered`, found %v", err)
	}
	if storeCalled != 1 || storeRolledBack != 1 {
		t.Fatalf("should try to store once then roll back")
	}
	if triedToSend != 1 {
		t.Fatalf("should try to send mail, once")
	}
}

func TestSendConfirmationGetEmailFailed(t *testing.T) {
	emailSenderMock := emailSender_mock.ServiceMock{}
	verificationMock := verification_mock.RepositoryMock{}
	authMock := auth_mock.RepositoryMock{
		GetUserInfoImpl: func(uid auth.UserId) (auth.UserInfo, error) {
			return auth.UserInfo{}, errors.New("some error")
		},
	}
	controller := defaultController.New(
		&verificationMock,
		&authMock,
		&emailSenderMock,
		logging.TestService(),
	)
	err := controller.SendConfirmationCode(verification.UserId(uuid.New().String()))
	if err == nil {
		t.Fatalf("`err` should not be nil")
	}
	if err.Code != verification.SendConfirmationCodeErrorInternal {
		t.Fatalf("unexpected error code, expected `internal`, found %v", err)
	}
}

func TestSendConfirmationCodeStoreFailed(t *testing.T) {
	storeCalled := 0
	storeRolledBack := 0
	emailSenderMock := emailSender_mock.ServiceMock{}
	verificationMock := verification_mock.RepositoryMock{
		StoreEmailVerificationCodeImpl: func(email string, code string) repositories.MutationWorkItem {
			return repositories.MutationWorkItem{
				Perform: func() error {
					storeCalled += 1
					return errors.New("some error")
				},
				Rollback: func() error {
					storeRolledBack += 1
					return nil
				},
			}
		},
	}
	authMock := auth_mock.RepositoryMock{
		GetUserInfoImpl: func(uid auth.UserId) (auth.UserInfo, error) {
			return auth.UserInfo{}, nil
		},
	}
	controller := defaultController.New(
		&verificationMock,
		&authMock,
		&emailSenderMock,
		logging.TestService(),
	)
	err := controller.SendConfirmationCode(verification.UserId(uuid.New().String()))
	if err == nil {
		t.Fatalf("`err` should not be nil")
	}
	if err.Code != verification.SendConfirmationCodeErrorInternal {
		t.Fatalf("unexpected error code, expected `internal`, found %v", err)
	}
	if storeCalled != 1 || storeRolledBack != 0 {
		t.Fatalf("should failed on store once then do not roll back")
	}
}

func TestSendConfirmationCodeDelivered(t *testing.T) {
	storeCalled := 0
	storeRolledBack := 0
	triedToSend := 0
	emailSenderMock := emailSender_mock.ServiceMock{
		SendImpl: func(subject, email string) error {
			triedToSend += 1
			return nil
		},
	}
	verificationMock := verification_mock.RepositoryMock{
		StoreEmailVerificationCodeImpl: func(email string, code string) repositories.MutationWorkItem {
			return repositories.MutationWorkItem{
				Perform: func() error {
					storeCalled += 1
					return nil
				},
				Rollback: func() error {
					storeRolledBack += 1
					return nil
				},
			}
		},
	}
	authMock := auth_mock.RepositoryMock{
		GetUserInfoImpl: func(uid auth.UserId) (auth.UserInfo, error) {
			return auth.UserInfo{}, nil
		},
	}
	controller := defaultController.New(
		&verificationMock,
		&authMock,
		&emailSenderMock,
		logging.TestService(),
	)
	err := controller.SendConfirmationCode(verification.UserId(uuid.New().String()))
	if err != nil {
		t.Fatalf("`err` should be nil, found %v", err)
	}
	if storeCalled != 1 || storeRolledBack != 0 {
		t.Fatalf("should store one then do not roll back")
	}
	if triedToSend != 1 {
		t.Fatalf("should send email, once")
	}
}

func TestConfirmEmailGetEmailFailed(t *testing.T) {
	emailSenderMock := emailSender_mock.ServiceMock{}
	verificationMock := verification_mock.RepositoryMock{}
	authMock := auth_mock.RepositoryMock{
		GetUserInfoImpl: func(uid auth.UserId) (auth.UserInfo, error) {
			return auth.UserInfo{}, errors.New("some error")
		},
	}
	controller := defaultController.New(
		&verificationMock,
		&authMock,
		&emailSenderMock,
		logging.TestService(),
	)
	err := controller.ConfirmEmail(verification.UserId(uuid.New().String()), "")
	if err == nil {
		t.Fatalf("`err` should not be nil")
	}
	if err.Code != verification.ConfirmEmailErrorInternal {
		t.Fatalf("unexpected error code, expected `internal`, found %v", err)
	}
}

func TestConfirmEmailGetCodeFailed(t *testing.T) {
	emailSenderMock := emailSender_mock.ServiceMock{}
	verificationMock := verification_mock.RepositoryMock{
		GetEmailVerificationCodeImpl: func(email string) (*string, error) {
			return nil, errors.New("some error")
		},
	}
	authMock := auth_mock.RepositoryMock{
		GetUserInfoImpl: func(uid auth.UserId) (auth.UserInfo, error) {
			return auth.UserInfo{}, nil
		},
	}
	controller := defaultController.New(
		&verificationMock,
		&authMock,
		&emailSenderMock,
		logging.TestService(),
	)
	err := controller.ConfirmEmail(verification.UserId(uuid.New().String()), "")
	if err == nil {
		t.Fatalf("`err` should not be nil")
	}
	if err.Code != verification.ConfirmEmailErrorInternal {
		t.Fatalf("unexpected error code, expected `internal`, found %v", err)
	}
}

func TestConfirmEmailCodeHasNotBeenSent(t *testing.T) {
	emailSenderMock := emailSender_mock.ServiceMock{}
	verificationMock := verification_mock.RepositoryMock{
		GetEmailVerificationCodeImpl: func(email string) (*string, error) {
			return nil, nil
		},
	}
	authMock := auth_mock.RepositoryMock{
		GetUserInfoImpl: func(uid auth.UserId) (auth.UserInfo, error) {
			return auth.UserInfo{}, nil
		},
	}
	controller := defaultController.New(
		&verificationMock,
		&authMock,
		&emailSenderMock,
		logging.TestService(),
	)
	err := controller.ConfirmEmail(verification.UserId(uuid.New().String()), "")
	if err == nil {
		t.Fatalf("`err` should not be nil")
	}
	if err.Code != verification.ConfirmEmailErrorCodeHasNotBeenSent {
		t.Fatalf("unexpected error code, expected `has not been sent`, found %v", err)
	}
}

func TestConfirmEmailCodeIsWrong(t *testing.T) {
	codeFromRepository := uuid.New().String()
	emailSenderMock := emailSender_mock.ServiceMock{}
	verificationMock := verification_mock.RepositoryMock{
		GetEmailVerificationCodeImpl: func(email string) (*string, error) {
			return &codeFromRepository, nil
		},
	}
	authMock := auth_mock.RepositoryMock{
		GetUserInfoImpl: func(uid auth.UserId) (auth.UserInfo, error) {
			return auth.UserInfo{}, nil
		},
	}
	controller := defaultController.New(
		&verificationMock,
		&authMock,
		&emailSenderMock,
		logging.TestService(),
	)
	err := controller.ConfirmEmail(verification.UserId(uuid.New().String()), "")
	if err == nil {
		t.Fatalf("`err` should not be nil")
	}
	if err.Code != verification.ConfirmEmailErrorWrongConfirmationCode {
		t.Fatalf("unexpected error code, expected `wrong code`, found %v", err)
	}
}

func TestConfirmEmailMarkValidatedFailed(t *testing.T) {
	codeFromRepository := uuid.New().String()
	emailSenderMock := emailSender_mock.ServiceMock{}
	verificationMock := verification_mock.RepositoryMock{
		GetEmailVerificationCodeImpl: func(email string) (*string, error) {
			return &codeFromRepository, nil
		},
	}
	authMock := auth_mock.RepositoryMock{
		GetUserInfoImpl: func(uid auth.UserId) (auth.UserInfo, error) {
			return auth.UserInfo{}, nil
		},
		MarkUserEmailValidatedImpl: func(uid auth.UserId) repositories.MutationWorkItem {
			return repositories.MutationWorkItem{
				Perform: func() error {
					return errors.New("some error")
				},
			}
		},
	}
	controller := defaultController.New(
		&verificationMock,
		&authMock,
		&emailSenderMock,
		logging.TestService(),
	)
	err := controller.ConfirmEmail(verification.UserId(uuid.New().String()), codeFromRepository)
	if err == nil {
		t.Fatalf("`err` should not be nil")
	}
	if err.Code != verification.ConfirmEmailErrorInternal {
		t.Fatalf("unexpected error code, expected `internal`, found %v", err)
	}
}

func TestConfirmEmailOk(t *testing.T) {
	codeFromRepository := uuid.New().String()
	emailSenderMock := emailSender_mock.ServiceMock{}
	verificationMock := verification_mock.RepositoryMock{
		GetEmailVerificationCodeImpl: func(email string) (*string, error) {
			return &codeFromRepository, nil
		},
	}
	authMock := auth_mock.RepositoryMock{
		GetUserInfoImpl: func(uid auth.UserId) (auth.UserInfo, error) {
			return auth.UserInfo{}, nil
		},
		MarkUserEmailValidatedImpl: func(uid auth.UserId) repositories.MutationWorkItem {
			return repositories.MutationWorkItem{
				Perform: func() error {
					return nil
				},
			}
		},
	}
	controller := defaultController.New(
		&verificationMock,
		&authMock,
		&emailSenderMock,
		logging.TestService(),
	)
	err := controller.ConfirmEmail(verification.UserId(uuid.New().String()), codeFromRepository)
	if err != nil {
		t.Fatalf("`err` should be nil, found %v", err)
	}
}
