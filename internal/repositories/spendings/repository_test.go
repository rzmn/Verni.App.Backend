package spendings_test

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"reflect"
	"sort"
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

func randomEid() spendings.ExpenseId {
	return spendings.ExpenseId(uuid.New().String())
}

func TestGetExpensesEmpty(t *testing.T) {
	repository := spendings.PostgresRepository(database)
	expenseId := randomEid()

	shouldBeEmpty, err := repository.GetExpense(expenseId)
	if err != nil {
		t.Fatalf("failed to get `shouldBeEmpty` err: %v", err)
	}
	if shouldBeEmpty != nil {
		t.Fatalf("`shouldBeEmpty` should be nil, found %v", *shouldBeEmpty)
	}
}

func TestGetBalanceEmpty(t *testing.T) {
	repository := spendings.PostgresRepository(database)
	counterparty := randomUid()

	shouldBeEmpty, err := repository.GetBalance(counterparty)
	if err != nil {
		t.Fatalf("failed to get `shouldBeEmpty` err: %v", err)
	}
	if len(shouldBeEmpty) != 0 {
		t.Fatalf("`shouldBeEmpty` should be empty, found %v", shouldBeEmpty)
	}
}

func TestExpensesAndCounterparties(t *testing.T) {
	repository := spendings.PostgresRepository(database)
	firstCounterparty := randomUid()
	secondCounterparty := randomUid()
	cost1 := spendings.Cost(456)
	cost2 := spendings.Cost(888)
	currency := spendings.Currency(uuid.New().String())

	expensesAreEqual := func(lhs spendings.Expense, rhs spendings.Expense) bool {
		sort.Slice(lhs.Shares, func(i, j int) bool {
			return lhs.Shares[i].Counterparty < lhs.Shares[j].Counterparty
		})
		sort.Slice(rhs.Shares, func(i, j int) bool {
			return rhs.Shares[i].Counterparty < rhs.Shares[j].Counterparty
		})
		return reflect.DeepEqual(lhs, rhs)
	}

	// adding first expense (created by first counterparty)
	// both first and second user are participants of that expense
	// checking an ability to get it

	firstExpense := spendings.Expense{
		Timestamp: 123,
		Details:   uuid.New().String(),
		Total:     cost1,
		Currency:  currency,
		Shares: []spendings.ShareOfExpense{
			{
				Counterparty: firstCounterparty,
				Cost:         cost1,
			},
			{
				Counterparty: secondCounterparty,
				Cost:         -cost1,
			},
		},
	}
	insertFirstExpenseTransaction := repository.AddExpense(firstExpense)
	firstExpenseId, err := insertFirstExpenseTransaction.Perform()
	if err != nil {
		t.Fatalf("failed to perform `insertFirstExpenseTransaction` err: %v", err)
	}
	shouldBeEqualToFirstExpense, err := repository.GetExpense(firstExpenseId)
	if err != nil {
		t.Fatalf("failed to get `shouldBeEqualToFirstExpense` err: %v", err)
	}
	if shouldBeEqualToFirstExpense == nil {
		t.Fatalf("`shouldBeEqualToFirstExpense` is nil, expected %v", firstExpense)
	}
	if !expensesAreEqual(shouldBeEqualToFirstExpense.Expense, firstExpense) {
		t.Fatalf("`shouldBeEqualToFirstExpense` is %v, expected %v", shouldBeEqualToFirstExpense.Expense, firstExpense)
	}

	// adding second expense (created by second counterparty). second expense's timestamp is bigger than the first's one
	// both first and second user are participants of that expense too
	// checking an ability to get it

	secondExpense := spendings.Expense{
		Timestamp: 1234,
		Details:   uuid.New().String(),
		Total:     cost2,
		Currency:  currency,
		Shares: []spendings.ShareOfExpense{
			{
				Counterparty: secondCounterparty,
				Cost:         -cost2 / 2,
			},
			{
				Counterparty: firstCounterparty,
				Cost:         cost2 / 2,
			},
		},
	}
	insertSecondExpenseTransaction := repository.AddExpense(secondExpense)
	secondExpenseId, err := insertSecondExpenseTransaction.Perform()
	if err != nil {
		t.Fatalf("failed to perform `insertSecondExpenseTransaction` err: %v", err)
	}
	shouldBeEqualToSecondExpense, err := repository.GetExpense(secondExpenseId)
	if err != nil {
		t.Fatalf("failed to get `shouldBeEqualToSecondExpense` err: %v", err)
	}
	if shouldBeEqualToSecondExpense == nil {
		t.Fatalf("`shouldBeEqualToSecondExpense` is nil, expected %v", secondExpense)
	}
	if !expensesAreEqual(shouldBeEqualToSecondExpense.Expense, secondExpense) {
		t.Fatalf("`shouldBeEqualToSecondExpense` is %v, expected %v", shouldBeEqualToSecondExpense.Expense, secondExpense)
	}

	// check that both expenses are available by GetExpensesBetween

	expensesBetween, err := repository.GetExpensesBetween(firstCounterparty, secondCounterparty)
	if err != nil {
		t.Fatalf("failed to get `expensesBetween` err: %v", err)
	}
	if len(expensesBetween) != 2 {
		t.Fatalf("should be 2 expenses in `expensesBetween`, found: %v", expensesBetween)
	}
	if !expensesAreEqual(expensesBetween[0].Expense, firstExpense) {
		t.Fatalf("first expense from `expensesBetween` should be equal to %v , found: %v", firstExpense, expensesBetween[0].Expense)
	}
	if !expensesAreEqual(expensesBetween[1].Expense, secondExpense) {
		t.Fatalf("first expense from `expensesBetween` should be equal to %v , found: %v", secondExpense, expensesBetween[1].Expense)
	}

	// check that first counterparty's balance mentions second counterparty

	firstCounterpartyBalance, err := repository.GetBalance(firstCounterparty)
	if err != nil {
		t.Fatalf("failed to get `firstCounterpartyBalance` err: %v", err)
	}
	if len(firstCounterpartyBalance) != 1 {
		t.Fatalf("`firstCounterpartyBalance` should contain second counterparty only, found %v", firstCounterpartyBalance)
	}
	if firstCounterpartyBalance[0].Counterparty != secondCounterparty ||
		len(firstCounterpartyBalance[0].Currencies) != 1 ||
		firstCounterpartyBalance[0].Currencies[currency] != (cost1+cost2/2) {
		t.Fatalf("`firstCounterpartyBalance` is incorrect %v", firstCounterpartyBalance[0])
	}

	// check that second counterparty's balance mentions second counterparty

	secondCounterpartyBalance, err := repository.GetBalance(secondCounterparty)
	if err != nil {
		t.Fatalf("failed to get `secondCounterpartyBalance` err: %v", err)
	}
	if len(secondCounterpartyBalance) != 1 {
		t.Fatalf("`secondCounterpartyBalance` should contain second counterparty only, found %v", secondCounterpartyBalance)
	}
	if secondCounterpartyBalance[0].Counterparty != firstCounterparty ||
		len(secondCounterpartyBalance[0].Currencies) != 1 ||
		secondCounterpartyBalance[0].Currencies[currency] != -(cost1+cost2/2) {
		t.Fatalf("`secondCounterpartyBalance` is incorrect %v", secondCounterpartyBalance[0])
	}

	// test second expense addition rollback works

	if err := insertSecondExpenseTransaction.Rollback(); err != nil {
		t.Fatalf("failed to rollback `insertSecondExpenseTransaction` err: %v", err)
	}
	expensesBetween, err = repository.GetExpensesBetween(firstCounterparty, secondCounterparty)
	if err != nil {
		t.Fatalf("[after first rollback] failed to get `expensesBetween` err: %v", err)
	}
	if len(expensesBetween) != 1 {
		t.Fatalf("[after first rollback] should be 1 expenses in `expensesBetween`, found: %v", expensesBetween)
	}
	if !expensesAreEqual(expensesBetween[0].Expense, firstExpense) {
		t.Fatalf("[after first rollback] first expense from `expensesBetween` should be equal to %v , found: %v", firstExpense, expensesBetween[0])
	}

	// test first expense addition rollback works

	if err := insertFirstExpenseTransaction.Rollback(); err != nil {
		t.Fatalf("failed to rollback `insertFirstExpenseTransaction` err: %v", err)
	}
	expensesBetween, err = repository.GetExpensesBetween(firstCounterparty, secondCounterparty)
	if err != nil {
		t.Fatalf("[after second rollback] failed to get `expensesBetween` err: %v", err)
	}
	if len(expensesBetween) != 0 {
		t.Fatalf("[after second rollback] should be 0 expenses in `expensesBetween`, found: %v", expensesBetween)
	}

	// check balances are empty after rollbacks

	firstCounterpartyBalance, err = repository.GetBalance(firstCounterparty)
	if err != nil {
		t.Fatalf("[after second rollback] failed to get `firstCounterpartyBalance` err: %v", err)
	}
	if len(firstCounterpartyBalance) != 0 {
		t.Fatalf("[after second rollback] `firstCounterpartyBalance` should be equal, found %v", firstCounterpartyBalance)
	}

	secondCounterpartyBalance, err = repository.GetBalance(secondCounterparty)
	if err != nil {
		t.Fatalf("[after second rollback] failed to get `secondCounterpartyBalance` err: %v", err)
	}
	if len(secondCounterpartyBalance) != 0 {
		t.Fatalf("[after second rollback] `secondCounterpartyBalance` should be empty, found %v", secondCounterpartyBalance)
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
