package spendings_test

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"testing"
	"verni/internal/common"
	"verni/internal/db"
	"verni/internal/repositories/spendings"

	"github.com/google/uuid"
)

var (
	database db.DB
)

func TestMain(m *testing.M) {
	database = func() db.DB {
		configFile, err := os.Open(common.AbsolutePath("./config/test/postgres_storage.json"))
		if err != nil {
			log.Fatalf("failed to open config file: %s", err)
		}
		defer configFile.Close()
		configData, err := io.ReadAll(configFile)
		if err != nil {
			log.Fatalf("failed to read config file: %s", err)
		}
		var config db.PostgresConfig
		json.Unmarshal([]byte(configData), &config)
		db, err := db.Postgres(config)
		if err != nil {
			log.Fatalf("failed to init db err: %v", err)
		}
		return db
	}()
	code := m.Run()

	os.Exit(code)
}

func init() {
	root, present := os.LookupEnv("VERNI_PROJECT_ROOT")
	if present {
		common.RegisterRelativePathRoot(root)
	} else {
		log.Fatalf("project root not found")
	}
}

func randomUid() spendings.CounterpartyId {
	return spendings.CounterpartyId(uuid.New().String())
}

func TestExpensesAndCounterparties(t *testing.T) {
	s := spendings.PostgresRepository(database)
	counterparty1 := randomUid()
	counterparty2 := randomUid()
	cost1 := spendings.Cost(456)
	cost2 := spendings.Cost(888)
	currency := spendings.Currency(uuid.New().String())

	expense1 := spendings.Expense{
		Timestamp: 123,
		Details:   uuid.New().String(),
		Total:     cost1,
		Currency:  currency,
		Shares: []spendings.ShareOfExpense{
			{
				Counterparty: counterparty1,
				Cost:         cost1,
			},
			{
				Counterparty: counterparty2,
				Cost:         -cost1,
			},
		},
	}
	insertTransaction := s.AddExpense(expense1)
	_, err := insertTransaction.Perform()
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	expense2 := spendings.Expense{
		Timestamp: 123,
		Details:   uuid.New().String(),
		Total:     cost2,
		Currency:  currency,
		Shares: []spendings.ShareOfExpense{
			{
				Counterparty: counterparty2,
				Cost:         -cost2 / 2,
			},
			{
				Counterparty: counterparty1,
				Cost:         cost2 / 2,
			},
		},
	}
	insertTransaction = s.AddExpense(expense2)
	_, err = insertTransaction.Perform()
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	expenses, err := s.GetExpensesBetween(counterparty1, counterparty2)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(expenses) != 2 {
		t.Fatalf("should be 2 expenses, found: %v", expenses)
	} else {
		log.Printf("expenses ok: %v\n", expenses)
	}
	counterparties, err := s.GetBalance(counterparty1)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(counterparties) != 1 || counterparties[0].Counterparty != counterparty2 || counterparties[0].Currencies[currency] != (cost1+cost2/2) {
		t.Fatalf("unexpected counterparty, found: %v", counterparties)
	}
	counterparties, err = s.GetBalance(counterparty2)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(counterparties) != 1 || counterparties[0].Counterparty != counterparty1 || counterparties[0].Currencies[currency] != -(cost1+cost2/2) {
		t.Fatalf("unexpected counterparty, found: %v", counterparties)
	}
}

func TestAddAndRemoveExpense(t *testing.T) {
	s := spendings.PostgresRepository(database)
	counterparty1 := randomUid()
	counterparty2 := randomUid()
	cost := spendings.Cost(456)
	currency := spendings.Currency(uuid.New().String())

	expense := spendings.Expense{
		Timestamp: 123,
		Details:   uuid.New().String(),
		Total:     cost,
		Currency:  currency,
		Shares: []spendings.ShareOfExpense{
			{
				Counterparty: counterparty1,
				Cost:         cost,
			},
			{
				Counterparty: counterparty2,
				Cost:         -cost,
			},
		},
	}
	insertTransaction := s.AddExpense(expense)
	_, err := insertTransaction.Perform()
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	expenses, err := s.GetExpensesBetween(counterparty1, counterparty2)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(expenses) != 1 {
		t.Fatalf("expenses len should be 1, found: %v", expenses)
	}
	expenseFromDb, err := s.GetExpense(expenses[0].Id)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if expenseFromDb == nil {
		t.Fatalf("deal should exists: %v", err)
	}
	deleteTransaction := s.RemoveExpense(expenses[0].Id)
	if err := deleteTransaction.Perform(); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	expenseFromDb, err = s.GetExpense(expenses[0].Id)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if expenseFromDb != nil {
		t.Fatalf("deal should not exists: %v", err)
	}
}
