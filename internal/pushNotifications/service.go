package pushNotifications

import (
	"encoding/json"
	"log"
	"os"
	"verni/internal/common"
	"verni/internal/storage"

	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/certificate"
)

type UserId storage.UserId
type Deal storage.IdentifiableDeal

type Service interface {
	FriendRequestHasBeenAccepted(receiver UserId, acceptedBy UserId)
	FriendRequestHasBeenReceived(receiver UserId, sentBy UserId)
	NewExpenseReceived(receiver UserId, deal Deal, author UserId)
}

type ApnsConfig struct {
	CertificatePath string `json:"certificatePath"`
	CredentialsPath string `json:"credentialsPath"`
}

func ApnsService(config ApnsConfig, db storage.Storage) (Service, error) {
	const op = "apns.AppleService"
	credentialsData, err := os.ReadFile(common.AbsolutePath(config.CredentialsPath))
	if err != nil {
		log.Printf("%s: failed to open config: %v", op, err)
		return &appleService{}, err
	}
	var credentials ApnsCredentials
	json.Unmarshal(credentialsData, &config)
	cert, err := certificate.FromP12File(common.AbsolutePath(config.CertificatePath), credentials.Password)
	if err != nil {
		log.Printf("%s: failed to open p12: %v", op, err)
		return &appleService{}, err
	}
	return &appleService{
		client:  apns2.NewClient(cert).Development(),
		storage: db,
	}, nil
}
