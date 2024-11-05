package pushNotifications

import (
	"encoding/json"
	"os"
	httpserver "verni/internal/http-server"
	pushNotificationsRepository "verni/internal/repositories/pushNotifications"
	"verni/internal/services/logging"
	"verni/internal/services/pathProvider"

	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/certificate"
)

type UserId httpserver.UserId
type Expense httpserver.IdentifiableExpense
type ExpenseId httpserver.ExpenseId
type Cost httpserver.Cost
type Repository pushNotificationsRepository.Repository

type Service interface {
	FriendRequestHasBeenAccepted(receiver UserId, acceptedBy UserId)
	FriendRequestHasBeenReceived(receiver UserId, sentBy UserId)
	NewExpenseReceived(receiver UserId, expense Expense, author UserId)
}

type ApnsConfig struct {
	CertificatePath string `json:"certificatePath"`
	CredentialsPath string `json:"credentialsPath"`
}

func ApnsService(config ApnsConfig, logger logging.Service, pathProviderService pathProvider.Service, repository Repository) (Service, error) {
	const op = "apns.AppleService"
	credentialsData, err := os.ReadFile(pathProviderService.AbsolutePath(config.CredentialsPath))
	if err != nil {
		logger.LogInfo("%s: failed to open config: %v", op, err)
		return &appleService{}, err
	}
	var credentials ApnsCredentials
	json.Unmarshal(credentialsData, &credentials)
	cert, err := certificate.FromP12File(pathProviderService.AbsolutePath(config.CertificatePath), credentials.Password)
	if err != nil {
		logger.LogInfo("%s: failed to open p12 creds %v: %v", op, err, credentials)
		return &appleService{}, err
	}
	return &appleService{
		client:     apns2.NewClient(cert).Development(),
		repository: repository,
		logger:     logger,
	}, nil
}
