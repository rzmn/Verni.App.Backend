package apns

import (
	"encoding/json"
	"log"
	"os"
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

func DefaultService(storage storage.Storage, certPath string, configPath string) (Service, error) {
	const op = "apns.PushNotificationSender.New"
	byteValue, err := os.ReadFile(configPath)
	if err != nil {
		log.Printf("%s: failed to open config: %v", op, err)
		return &defaultService{}, err
	}
	var config Config
	json.Unmarshal(byteValue, &config)
	cert, err := certificate.FromP12File(certPath, config.Password)
	if err != nil {
		log.Printf("%s: failed to open p12: %v", op, err)
		return &defaultService{}, err
	}
	return &defaultService{
		client:  apns2.NewClient(cert).Development(),
		storage: storage,
	}, nil
}
