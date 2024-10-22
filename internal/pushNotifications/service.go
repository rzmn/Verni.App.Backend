package pushNotifications

import (
	"encoding/json"
	"log"
	"os"
	"verni/internal/common"
	pushNotificationsRepository "verni/internal/repositories/pushNotifications"
	spendingsRepository "verni/internal/repositories/spendings"
	"verni/internal/storage"

	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/certificate"
)

type UserId storage.UserId
type Expense spendingsRepository.IdentifiableExpense
type Repository pushNotificationsRepository.Repository

type Service interface {
	FriendRequestHasBeenAccepted(receiver UserId, acceptedBy UserId)
	FriendRequestHasBeenReceived(receiver UserId, sentBy UserId)
	NewExpenseReceived(receiver UserId, deal Expense, author UserId)
}

type ApnsConfig struct {
	CertificatePath string `json:"certificatePath"`
	CredentialsPath string `json:"credentialsPath"`
}

func ApnsService(config ApnsConfig, repository Repository) (Service, error) {
	const op = "apns.AppleService"
	credentialsData, err := os.ReadFile(common.AbsolutePath(config.CredentialsPath))
	if err != nil {
		log.Printf("%s: failed to open config: %v", op, err)
		return &appleService{}, err
	}
	var credentials ApnsCredentials
	json.Unmarshal(credentialsData, &credentials)
	cert, err := certificate.FromP12File(common.AbsolutePath(config.CertificatePath), credentials.Password)
	if err != nil {
		log.Printf("%s: failed to open p12 creds %v: %v", op, err, credentials)
		return &appleService{}, err
	}
	return &appleService{
		client:     apns2.NewClient(cert).Development(),
		repository: repository,
	}, nil
}
