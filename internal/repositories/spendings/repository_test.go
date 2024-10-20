package spendings_test

import (
	"log"
	"testing"
	"verni/internal/repositories/spendings"
	"verni/internal/storage"

	"github.com/google/uuid"
)

var (
	_s *spendings.Repository
)

func getRepository(t *testing.T) spendings.Repository {
	if _s != nil {
		return *_s
	}
	repository, err := spendings.PostgresRepository(
		spendings.PostgresConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "tester",
			Password: "test_password",
			DbName:   "mydb",
		},
	)
	if err != nil {
		t.Fatalf("failed to init repository err: %v", err)
	}
	_s = &repository
	return repository
}

func randomUid() spendings.UserId {
	return spendings.UserId(uuid.New().String())
}

func TestDealsAndCounterparties(t *testing.T) {
	s := getRepository(t)
	counterparty1 := randomUid()
	counterparty2 := randomUid()
	cost1 := int64(456)
	cost2 := int64(888)
	currency := uuid.New().String()

	deal1 := spendings.Deal{
		Timestamp: 123,
		Details:   uuid.New().String(),
		Cost:      cost1,
		Currency:  currency,
		Spendings: []storage.Spending{
			{
				UserId: storage.UserId(counterparty1),
				Cost:   cost1,
			},
			{
				UserId: storage.UserId(counterparty2),
				Cost:   -cost1,
			},
		},
	}
	insertTransaction := s.InsertDeal(deal1)
	_, err := insertTransaction.Perform()
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	deal2 := storage.Deal{
		Timestamp: 123,
		Details:   uuid.New().String(),
		Cost:      cost2,
		Currency:  currency,
		Spendings: []storage.Spending{
			{
				UserId: storage.UserId(counterparty2),
				Cost:   -cost2 / 2,
			},
			{
				UserId: storage.UserId(counterparty1),
				Cost:   cost2 / 2,
			},
		},
	}
	insertTransaction = s.InsertDeal(spendings.Deal(deal2))
	_, err = insertTransaction.Perform()
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	deals, err := s.GetDeals(counterparty1, counterparty2)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(deals) != 2 {
		t.Fatalf("should be 2 deals, found: %v", deals)
	} else {
		log.Printf("deals ok: %v\n", deals)
	}
	counterparties, err := s.GetCounterparties(counterparty1)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(counterparties) != 1 || counterparties[0].Counterparty != string(counterparty2) || counterparties[0].Balance[currency] != (cost1+cost2/2) {
		t.Fatalf("unexpected counterparty, found: %v", counterparties)
	}
	counterparties, err = s.GetCounterparties(counterparty2)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(counterparties) != 1 || counterparties[0].Counterparty != string(counterparty1) || counterparties[0].Balance[currency] != -(cost1+cost2/2) {
		t.Fatalf("unexpected counterparty, found: %v", counterparties)
	}
}

func TestInsertAndRemoveDeal(t *testing.T) {
	s := getRepository(t)
	counterparty1 := randomUid()
	counterparty2 := randomUid()
	cost := int64(456)
	currency := uuid.New().String()

	deal := storage.Deal{
		Timestamp: 123,
		Details:   uuid.New().String(),
		Cost:      cost,
		Currency:  currency,
		Spendings: []storage.Spending{
			{
				UserId: storage.UserId(counterparty1),
				Cost:   cost,
			},
			{
				UserId: storage.UserId(counterparty2),
				Cost:   -cost,
			},
		},
	}
	insertTransaction := s.InsertDeal(spendings.Deal(deal))
	_, err := insertTransaction.Perform()
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	deals, err := s.GetDeals(counterparty1, counterparty2)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(deals) != 1 {
		t.Fatalf("deals len should be 1, found: %v", deals)
	}
	dealFromDb, err := s.GetDeal(spendings.DealId(deals[0].Id))
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if dealFromDb == nil {
		t.Fatalf("deal should exists: %v", err)
	}
	deleteTransaction := s.RemoveDeal(spendings.DealId(deals[0].Id))
	if err := deleteTransaction.Perform(); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	dealFromDb, err = s.GetDeal(spendings.DealId(deals[0].Id))
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if dealFromDb != nil {
		t.Fatalf("deal should not exists: %v", err)
	}
}

func TestGetCounterpartiesForDeal(t *testing.T) {
	s := getRepository(t)
	counterparty1 := randomUid()
	counterparty2 := randomUid()
	cost := int64(456)
	currency := uuid.New().String()

	deal := storage.Deal{
		Timestamp: 123,
		Details:   uuid.New().String(),
		Cost:      cost,
		Currency:  currency,
		Spendings: []storage.Spending{
			{
				UserId: storage.UserId(counterparty1),
				Cost:   cost,
			},
			{
				UserId: storage.UserId(counterparty2),
				Cost:   -cost,
			},
		},
	}
	insertTransaction := s.InsertDeal(spendings.Deal(deal))
	_, err := insertTransaction.Perform()
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	deals, err := s.GetDeals(counterparty1, counterparty2)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(deals) != 1 {
		t.Fatalf("deals len should be 1, found: %v", deals)
	}
	counterparties, err := s.GetCounterpartiesForDeal(spendings.DealId(deals[0].Id))
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(counterparties) != 2 {
		t.Fatalf("counterparties len should be 2, found: %v", counterparties)
	}
	passedUsers := map[spendings.UserId]bool{}
	for i := 0; i < len(counterparties); i++ {
		if counterparties[i] == counterparty1 {
			passedUsers[counterparties[i]] = true
		}
		if counterparties[i] == counterparty2 {
			passedUsers[counterparties[i]] = true
		}
	}
	if len(passedUsers) != 2 {
		t.Fatalf("passedUsers len should be 2, found: %v", passedUsers)
	}
}
