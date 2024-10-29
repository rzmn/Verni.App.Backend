package verification_test

import (
	"errors"
	"testing"
	"verni/internal/controllers/verification"
	"verni/internal/repositories"
	"verni/internal/repositories/auth"
	auth_mock "verni/internal/repositories/auth/mock"
	verification_mock "verni/internal/repositories/verification/mock"
	emailSender_mock "verni/internal/services/emailSender/mock"

	"github.com/google/uuid"
)

func TestSendConfirmationCodeNotDelivered(t *testing.T) {
	emailSenderMock := emailSender_mock.ServiceMock{
		SendImpl: func(subject, email string) error {
			return errors.New("some error")
		},
	}
	verificationMock := verification_mock.RepositoryMock{
		StoreEmailVerificationCodeImpl: func(email string, code string) repositories.MutationWorkItem {
			return repositories.MutationWorkItem{
				Perform:  func() error { return nil },
				Rollback: func() error { return nil },
			}
		},
	}
	authMock := auth_mock.RepositoryMock{
		GetUserInfoImpl: func(uid auth.UserId) (auth.UserInfo, error) {
			return auth.UserInfo{}, nil
		},
	}
	controller := verification.DefaultController(
		&verificationMock,
		&authMock,
		&emailSenderMock,
	)
	err := controller.SendConfirmationCode(verification.UserId(uuid.New().String()))
	if err == nil {
		t.Fatalf("`err` should not be nil")
	}
	if err.Code != verification.SendConfirmationCodeErrorNotDelivered {
		t.Fatalf("unexpected error code, expected `not delivered` found %v", err)
	}
}
