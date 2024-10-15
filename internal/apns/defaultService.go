package apns

import (
	"accounty/internal/storage"
	"encoding/json"
	"fmt"
	"log"

	"github.com/sideshow/apns2"
)

type defaultService struct {
	client  *apns2.Client
	storage storage.Storage
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

func (s *defaultService) FriendRequestHasBeenAccepted(receiver UserId, acceptedBy UserId) {
	const op = "apns.defaultService.FriendRequestHasBeenAccepted"
	log.Printf("%s: start[receiver=%s acceptedBy=%s]", op, receiver, acceptedBy)
	receiverToken, err := s.storage.GetPushToken(storage.UserId(receiver))
	if err != nil {
		log.Printf("%s: cannot get receiver token from db err: %v", op, err)
		return
	}
	if receiverToken == nil {
		log.Printf("%s: receiver push token is nil", op)
		return
	}
	type Payload struct {
		Target UserId `json:"t"`
	}
	body := fmt.Sprintf("By %s", acceptedBy)
	payload := Payload{
		Target: acceptedBy,
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
		log.Printf("%s: failed to create payload string: %v", op, err)
		return
	}
	if err := s.send(*receiverToken, string(payloadString)); err != nil {
		log.Printf("%s: failed to send push: %v", op, err)
		return
	}
	log.Printf("%s: success[receiver=%s acceptedBy=%s]", op, receiver, acceptedBy)
}

func (s *defaultService) FriendRequestHasBeenReceived(receiver UserId, sentBy UserId) {
	const op = "apns.defaultService.FriendRequestHasBeenReceived"
	log.Printf("%s: start[receiver=%s sentBy=%s]", op, receiver, sentBy)
	receiverToken, err := s.storage.GetPushToken(storage.UserId(receiver))
	if err != nil {
		log.Printf("%s: cannot get receiver token from db err: %v", op, err)
		return
	}
	if receiverToken == nil {
		log.Printf("%s: receiver push token is nil", op)
		return
	}
	type Payload struct {
		Sender UserId `json:"s"`
	}
	body := fmt.Sprintf("From: %s", sentBy)
	payload := Payload{
		Sender: sentBy,
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
		log.Printf("%s: failed to create payload string: %v", op, err)
		return
	}
	if err := s.send(*receiverToken, string(payloadString)); err != nil {
		log.Printf("%s: failed to send push: %v", op, err)
		return
	}
	log.Printf("%s: success[receiver=%s sentBy=%s]", op, receiver, sentBy)
}

func (s *defaultService) NewExpenseReceived(receiver UserId, deal Deal, author UserId) {
	const op = "apns.defaultService.NewExpenseReceived"
	log.Printf("%s: start[receiver=%s did=%s author=%s]", op, receiver, deal.Id, author)
	receiverToken, err := s.storage.GetPushToken(storage.UserId(receiver))
	if err != nil {
		log.Printf("%s: cannot get receiver token from db err: %v", op, err)
		return
	}
	if receiverToken == nil {
		log.Printf("%s: receiver push token is nil", op)
		return
	}
	type Payload struct {
		DealId   storage.DealId `json:"d"`
		AuthorId UserId         `json:"u"`
		Cost     int64          `json:"c"`
	}
	body := fmt.Sprintf("%s: %d", deal.Details, deal.Cost)
	cost := deal.Cost
	for i := 0; i < len(deal.Spendings); i++ {
		if deal.Spendings[i].UserId == storage.UserId(receiver) {
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
	if err := s.send(*receiverToken, string(payloadString)); err != nil {
		log.Printf("%s: failed to send push: %v", op, err)
		return
	}
	log.Printf("%s: success[receiver=%s did=%s author=%s]", op, receiver, deal.Id, author)
}

func (s *defaultService) send(token string, payloadString string) error {
	const op = "apns.defaultService.send"
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
