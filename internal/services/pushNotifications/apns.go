package pushNotifications

import (
	"encoding/json"
	"fmt"
	"log"
	"verni/internal/repositories/pushNotifications"

	"github.com/sideshow/apns2"
)

type appleService struct {
	client     *apns2.Client
	repository Repository
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

type ApnsCredentials struct {
	Password string `json:"cert_pwd"`
}

func (s *appleService) FriendRequestHasBeenAccepted(receiver UserId, acceptedBy UserId) {
	const op = "apns.defaultService.FriendRequestHasBeenAccepted"
	log.Printf("%s: start[receiver=%s acceptedBy=%s]", op, receiver, acceptedBy)
	receiverToken, err := s.repository.GetPushToken(pushNotifications.UserId(receiver))
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

func (s *appleService) FriendRequestHasBeenReceived(receiver UserId, sentBy UserId) {
	const op = "apns.defaultService.FriendRequestHasBeenReceived"
	log.Printf("%s: start[receiver=%s sentBy=%s]", op, receiver, sentBy)
	receiverToken, err := s.repository.GetPushToken(pushNotifications.UserId(receiver))
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

func (s *appleService) NewExpenseReceived(receiver UserId, expense Expense, author UserId) {
	const op = "apns.defaultService.NewExpenseReceived"
	log.Printf("%s: start[receiver=%s id=%s author=%s]", op, receiver, expense.Id, author)
	receiverToken, err := s.repository.GetPushToken(pushNotifications.UserId(receiver))
	if err != nil {
		log.Printf("%s: cannot get receiver token from db err: %v", op, err)
		return
	}
	if receiverToken == nil {
		log.Printf("%s: receiver push token is nil", op)
		return
	}
	type Payload struct {
		DealId   ExpenseId `json:"d"`
		AuthorId UserId    `json:"u"`
		Cost     Cost      `json:"c"`
	}
	body := fmt.Sprintf("%s: %d", expense.Details, expense.Total)
	cost := expense.Total
	for i := 0; i < len(expense.Shares); i++ {
		if UserId(expense.Shares[i].UserId) == receiver {
			cost = expense.Shares[i].Cost
		}
	}
	payload := Payload{
		DealId:   ExpenseId(expense.Id),
		AuthorId: author,
		Cost:     Cost(cost),
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
	log.Printf("%s: success[receiver=%s id=%s author=%s]", op, receiver, expense.Id, author)
}

func (s *appleService) send(token string, payloadString string) error {
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
