package apns

import (
	"accounty/internal/storage"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/certificate"
)

type PushNotificationSender struct {
	client *apns2.Client
}

type PushDataType int

const (
	PushDataTypeFriendRequestHasBeenAccepted = iota
	PushDataTypeGotFriendRequest
	PushDataTypeNewExpenseReceived
)

type PushData[T any] struct {
	Type    PushDataType `json:"t"`
	Payload *T           `json:"p,omitempty"`
}

type Push[T any] struct {
	Aps  PushPayload `json:"aps"`
	Data PushData[T] `json:"d"`
}

type PushPayload struct {
	MutableContent *int             `json:"mutable-content,omitempty"`
	Alert          PushPayloadAlert `json:"alert"`
}

type PushPayloadAlert struct {
	Title    string  `json:"title"`
	Subtitle *string `json:"subtitle,omitempty"`
	Body     *string `json:"body,omitempty"`
}

type Config struct {
	Password string `json:"cert_pwd"`
}

func (s *PushNotificationSender) FriendRequestHasBeenAccepted(token string, target storage.UserId) {
	const op = "apns.PushNotificationSender.FriendRequestHasBeenAccepted"
	type Payload struct {
		Target storage.UserId `json:"t"`
	}
	body := fmt.Sprintf("From: %s", target)
	payload := Payload{
		Target: target,
	}
	mutable := 1
	payloadString, err := json.Marshal(Push[Payload]{
		Aps: PushPayload{
			MutableContent: &mutable,
			Alert: PushPayloadAlert{
				Title:    "Friend request has been accepted",
				Subtitle: nil,
				Body:     &body,
			},
		},
		Data: PushData[Payload]{
			Type:    PushDataTypeFriendRequestHasBeenAccepted,
			Payload: &payload,
		},
	})
	if err != nil {
		log.Printf("%s: failed create payload string: %v", op, err)
		return
	}
	if err := s.send(token, string(payloadString)); err != nil {
		log.Printf("%s: failed to send push: %v", op, err)
		return
	}
}

func (s *PushNotificationSender) GotFriendRequest(token string, sender storage.UserId) {
	const op = "apns.PushNotificationSender.GotFriendRequest"
	type Payload struct {
		Sender storage.UserId `json:"s"`
	}
	body := fmt.Sprintf("From: %s", sender)
	payload := Payload{
		Sender: sender,
	}
	mutable := 1
	payloadString, err := json.Marshal(Push[Payload]{
		Aps: PushPayload{
			MutableContent: &mutable,
			Alert: PushPayloadAlert{
				Title:    "Got Friend Request",
				Subtitle: nil,
				Body:     &body,
			},
		},
		Data: PushData[Payload]{
			Type:    PushDataTypeGotFriendRequest,
			Payload: &payload,
		},
	})
	if err != nil {
		log.Printf("%s: failed create payload string: %v", op, err)
		return
	}
	if err := s.send(token, string(payloadString)); err != nil {
		log.Printf("%s: failed to send push: %v", op, err)
		return
	}
}

func (s *PushNotificationSender) NewExpenseReceived(token string, deal storage.IdentifiableDeal, author storage.UserId, receiver storage.UserId) {
	const op = "apns.PushNotificationSender.NewExpenseReceived"
	type Payload struct {
		DealId   storage.DealId `json:"d"`
		AuthorId storage.UserId `json:"u"`
		Cost     int64          `json:"c"`
	}
	body := fmt.Sprintf("%s: %d", deal.Details, deal.Cost)
	cost := deal.Cost
	for i := 0; i < len(deal.Spendings); i++ {
		if deal.Spendings[i].UserId == receiver {
			cost = deal.Spendings[i].Cost
		}
	}
	payload := Payload{
		DealId:   deal.Id,
		AuthorId: author,
		Cost:     cost,
	}
	mutable := 1
	payloadString, err := json.Marshal(Push[Payload]{
		Aps: PushPayload{
			MutableContent: &mutable,
			Alert: PushPayloadAlert{
				Title:    "New Expense Received",
				Subtitle: nil,
				Body:     &body,
			},
		},
		Data: PushData[Payload]{
			Type:    PushDataTypeNewExpenseReceived,
			Payload: &payload,
		},
	})
	if err != nil {
		log.Printf("%s: failed create payload string: %v", op, err)
		return
	}
	if err := s.send(token, string(payloadString)); err != nil {
		log.Printf("%s: failed to send push: %v", op, err)
		return
	}
}

func (s *PushNotificationSender) send(token string, payloadString string) error {
	const op = "apns.PushNotificationSender.Send"
	notification := &apns2.Notification{}
	notification.DeviceToken = token
	notification.Topic = "com.rzmn.accountydev.app"

	log.Printf("%s: sending push: %s", op, payloadString)
	notification.Payload = payloadString

	res, err := s.client.Push(notification)

	if err != nil {
		log.Printf("%s: failed to send notification: %v", op, err)
		return err
	}
	fmt.Printf("%s: sent %v %v %v\n", op, res.StatusCode, res.ApnsID, res.Reason)
	return nil
}

func New(certPath string, configPath string) (PushNotificationSender, error) {
	const op = "apns.PushNotificationSender.New"
	byteValue, err := os.ReadFile(configPath)
	if err != nil {
		log.Printf("%s: failed to open config: %v", op, err)
		return PushNotificationSender{}, err
	}
	var config Config
	json.Unmarshal(byteValue, &config)
	cert, err := certificate.FromP12File(certPath, config.Password)
	if err != nil {
		log.Printf("%s: failed to open p12: %v", op, err)
		return PushNotificationSender{}, err
	}
	return PushNotificationSender{
		client: apns2.NewClient(cert).Development(),
	}, nil
}
