package spendings_test

import (
	"errors"
	"testing"
	"verni/internal/controllers/spendings"
	"verni/internal/repositories"
	spendingsRepository "verni/internal/repositories/spendings"
	spendings_mock "verni/internal/repositories/spendings/mock"
	"verni/internal/services/pushNotifications"
	pushNotifications_mock "verni/internal/services/pushNotifications/mock"

	"github.com/google/uuid"
)

func TestAddExpenseFailedNotYourExpense(t *testing.T) {
	repository := spendings_mock.RepositoryMock{}
	pushService := pushNotifications_mock.ServiceMock{}

	controller := spendings.DefaultController(&repository, &pushService)

	expense := spendings.Expense{
		Shares: []spendingsRepository.ShareOfExpense{
			{
				Counterparty: spendingsRepository.CounterpartyId(uuid.New().String()),
			},
			{
				Counterparty: spendingsRepository.CounterpartyId(uuid.New().String()),
			},
		},
	}
	err := controller.AddExpense(expense, spendings.CounterpartyId(uuid.New().String()))
	if err == nil {
		t.Fatalf("`AddExpense` should be failed, found nil err")
	}
	if err.Code != spendings.CreateDealErrorNotYourExpense {
		t.Fatalf("`AddExpense` should be failed with `not your expese`, found err %v", err)
	}
}

func TestAddExpenseFailedToAddInRepository(t *testing.T) {
	repository := spendings_mock.RepositoryMock{
		AddExpenseImpl: func(id spendingsRepository.Expense) repositories.MutationWorkItemWithReturnValue[spendingsRepository.ExpenseId] {
			return repositories.MutationWorkItemWithReturnValue[spendingsRepository.ExpenseId]{
				Perform: func() (spendingsRepository.ExpenseId, error) {
					return spendingsRepository.ExpenseId(uuid.New().String()), errors.New("some error")
				},
			}
		},
	}
	pushService := pushNotifications_mock.ServiceMock{}

	controller := spendings.DefaultController(&repository, &pushService)
	actor := spendings.CounterpartyId(uuid.New().String())
	counterparty := spendings.CounterpartyId(uuid.New().String())

	expense := spendings.Expense{
		Shares: []spendingsRepository.ShareOfExpense{
			{
				Counterparty: spendingsRepository.CounterpartyId(actor),
			},
			{
				Counterparty: spendingsRepository.CounterpartyId(counterparty),
			},
		},
	}
	err := controller.AddExpense(expense, actor)
	if err == nil {
		t.Fatalf("`AddExpense` should be failed, found nil err")
	}
	if err.Code != spendings.CreateDealErrorInternal {
		t.Fatalf("`AddExpense` should be failed with `internal`, found err %v", err)
	}
}

func TestAddExpenseOk(t *testing.T) {
	repository := spendings_mock.RepositoryMock{
		AddExpenseImpl: func(id spendingsRepository.Expense) repositories.MutationWorkItemWithReturnValue[spendingsRepository.ExpenseId] {
			return repositories.MutationWorkItemWithReturnValue[spendingsRepository.ExpenseId]{
				Perform: func() (spendingsRepository.ExpenseId, error) {
					return spendingsRepository.ExpenseId(uuid.New().String()), nil
				},
			}
		},
	}
	pushSentTo := []spendings.CounterpartyId{}
	pushService := pushNotifications_mock.ServiceMock{
		NewExpenseReceivedImpl: func(receiver pushNotifications.UserId, expense pushNotifications.Expense, author pushNotifications.UserId) {
			pushSentTo = append(pushSentTo, spendings.CounterpartyId(receiver))
		},
	}

	controller := spendings.DefaultController(&repository, &pushService)
	actor := spendings.CounterpartyId(uuid.New().String())
	counterparty := spendings.CounterpartyId(uuid.New().String())

	expense := spendings.Expense{
		Shares: []spendingsRepository.ShareOfExpense{
			{
				Counterparty: spendingsRepository.CounterpartyId(actor),
			},
			{
				Counterparty: spendingsRepository.CounterpartyId(counterparty),
			},
		},
	}
	err := controller.AddExpense(expense, actor)
	if err != nil {
		t.Fatalf("`AddExpense` should not be failed, found err %v", err)
	}
	if len(pushSentTo) != 1 || pushSentTo[0] != counterparty {
		t.Fatalf("`AddExpense` should trigger push to %s, found %v", counterparty, pushSentTo)
	}
}

func TestRemoveExpenseFailedToGetById(t *testing.T) {
	repository := spendings_mock.RepositoryMock{
		GetExpenseImpl: func(id spendingsRepository.ExpenseId) (*spendingsRepository.IdentifiableExpense, error) {
			return nil, errors.New("some error")
		},
	}
	pushService := pushNotifications_mock.ServiceMock{}
	controller := spendings.DefaultController(&repository, &pushService)
	_, err := controller.RemoveExpense(spendings.ExpenseId(uuid.New().String()), spendings.CounterpartyId(uuid.New().String()))
	if err == nil {
		t.Fatalf("`RemoveExpense` should be failed, found nil err")
	}
	if err.Code != spendings.DeleteDealErrorInternal {
		t.Fatalf("`RemoveExpense` should be failed with `internal`, found err %v", err)
	}
}

func TestRemoveExpenseFailedNotFound(t *testing.T) {
	repository := spendings_mock.RepositoryMock{
		GetExpenseImpl: func(id spendingsRepository.ExpenseId) (*spendingsRepository.IdentifiableExpense, error) {
			return nil, nil
		},
	}
	pushService := pushNotifications_mock.ServiceMock{}
	controller := spendings.DefaultController(&repository, &pushService)
	_, err := controller.RemoveExpense(spendings.ExpenseId(uuid.New().String()), spendings.CounterpartyId(uuid.New().String()))
	if err == nil {
		t.Fatalf("`RemoveExpense` should be failed, found nil err")
	}
	if err.Code != spendings.DeleteDealErrorDealNotFound {
		t.Fatalf("`RemoveExpense` should be failed with `not found`, found err %v", err)
	}
}

func TestRemoveExpenseFailedNotYourExpense(t *testing.T) {
	repository := spendings_mock.RepositoryMock{
		GetExpenseImpl: func(id spendingsRepository.ExpenseId) (*spendingsRepository.IdentifiableExpense, error) {
			return &spendingsRepository.IdentifiableExpense{
				Id: id,
				Expense: spendingsRepository.Expense{
					Shares: []spendingsRepository.ShareOfExpense{
						{
							Counterparty: spendingsRepository.CounterpartyId(uuid.New().String()),
						},
						{
							Counterparty: spendingsRepository.CounterpartyId(uuid.New().String()),
						},
					},
				},
			}, nil
		},
	}
	pushService := pushNotifications_mock.ServiceMock{}
	controller := spendings.DefaultController(&repository, &pushService)
	_, err := controller.RemoveExpense(spendings.ExpenseId(uuid.New().String()), spendings.CounterpartyId(uuid.New().String()))
	if err == nil {
		t.Fatalf("`RemoveExpense` should be failed, found nil err")
	}
	if err.Code != spendings.DeleteDealErrorNotYourExpense {
		t.Fatalf("`RemoveExpense` should be failed with `not your expense`, found err %v", err)
	}
}

func TestRemoveExpenseRemoveFailed(t *testing.T) {
	actor := spendings.CounterpartyId(uuid.New().String())
	counterparty := spendings.CounterpartyId(uuid.New().String())
	repository := spendings_mock.RepositoryMock{
		GetExpenseImpl: func(id spendingsRepository.ExpenseId) (*spendingsRepository.IdentifiableExpense, error) {
			return &spendingsRepository.IdentifiableExpense{
				Id: id,
				Expense: spendingsRepository.Expense{
					Shares: []spendingsRepository.ShareOfExpense{
						{
							Counterparty: spendingsRepository.CounterpartyId(actor),
						},
						{
							Counterparty: spendingsRepository.CounterpartyId(counterparty),
						},
					},
				},
			}, nil
		},
		RemoveExpenseImpl: func(id spendingsRepository.ExpenseId) repositories.MutationWorkItem {
			return repositories.MutationWorkItem{
				Perform: func() error {
					return errors.New("some error")
				},
			}
		},
	}
	pushService := pushNotifications_mock.ServiceMock{}
	controller := spendings.DefaultController(&repository, &pushService)
	_, err := controller.RemoveExpense(spendings.ExpenseId(uuid.New().String()), actor)
	if err == nil {
		t.Fatalf("`RemoveExpense` should be failed, found nil err")
	}
	if err.Code != spendings.DeleteDealErrorInternal {
		t.Fatalf("`RemoveExpense` should be failed with `internal`, found err %v", err)
	}
}

func TestRemoveExpenseOk(t *testing.T) {
	removeCalls := 0
	actor := spendings.CounterpartyId(uuid.New().String())
	counterparty := spendings.CounterpartyId(uuid.New().String())
	repository := spendings_mock.RepositoryMock{
		GetExpenseImpl: func(id spendingsRepository.ExpenseId) (*spendingsRepository.IdentifiableExpense, error) {
			return &spendingsRepository.IdentifiableExpense{
				Id: id,
				Expense: spendingsRepository.Expense{
					Shares: []spendingsRepository.ShareOfExpense{
						{
							Counterparty: spendingsRepository.CounterpartyId(actor),
						},
						{
							Counterparty: spendingsRepository.CounterpartyId(counterparty),
						},
					},
				},
			}, nil
		},
		RemoveExpenseImpl: func(id spendingsRepository.ExpenseId) repositories.MutationWorkItem {
			return repositories.MutationWorkItem{
				Perform: func() error {
					removeCalls += 1
					return nil
				},
			}
		},
	}
	pushService := pushNotifications_mock.ServiceMock{}
	controller := spendings.DefaultController(&repository, &pushService)
	_, err := controller.RemoveExpense(spendings.ExpenseId(uuid.New().String()), actor)
	if err != nil {
		t.Fatalf("`RemoveExpense` should not be failed, found err %v", err)
	}
	if removeCalls != 1 {
		t.Fatalf("remove should be called once, found %d", removeCalls)
	}
}

func TestGetExpenseFailedToGetFromRepository(t *testing.T) {
	repository := spendings_mock.RepositoryMock{
		GetExpenseImpl: func(id spendingsRepository.ExpenseId) (*spendingsRepository.IdentifiableExpense, error) {
			return nil, errors.New("some error")
		},
	}
	pushService := pushNotifications_mock.ServiceMock{}
	controller := spendings.DefaultController(&repository, &pushService)
	_, err := controller.GetExpense(spendings.ExpenseId(uuid.New().String()), spendings.CounterpartyId(uuid.New().String()))
	if err == nil {
		t.Fatalf("`GetExpense` should be failed, found nil err")
	}
	if err.Code != spendings.GetDealErrorInternal {
		t.Fatalf("`GetExpense` should be failed with `internal`, found err %v", err)
	}
}

func TestGetExpenseFailedNotFound(t *testing.T) {
	repository := spendings_mock.RepositoryMock{
		GetExpenseImpl: func(id spendingsRepository.ExpenseId) (*spendingsRepository.IdentifiableExpense, error) {
			return nil, nil
		},
	}
	pushService := pushNotifications_mock.ServiceMock{}
	controller := spendings.DefaultController(&repository, &pushService)
	_, err := controller.GetExpense(spendings.ExpenseId(uuid.New().String()), spendings.CounterpartyId(uuid.New().String()))
	if err == nil {
		t.Fatalf("`GetExpense` should be failed, found nil err")
	}
	if err.Code != spendings.GetDealErrorDealNotFound {
		t.Fatalf("`GetExpense` should be failed with `not found`, found err %v", err)
	}
}

func TestGetExpenseNotYourExpense(t *testing.T) {
	repository := spendings_mock.RepositoryMock{
		GetExpenseImpl: func(id spendingsRepository.ExpenseId) (*spendingsRepository.IdentifiableExpense, error) {
			return &spendingsRepository.IdentifiableExpense{
				Id: id,
				Expense: spendingsRepository.Expense{
					Shares: []spendingsRepository.ShareOfExpense{
						{
							Counterparty: spendingsRepository.CounterpartyId(uuid.New().String()),
						},
						{
							Counterparty: spendingsRepository.CounterpartyId(uuid.New().String()),
						},
					},
				},
			}, nil
		},
	}
	pushService := pushNotifications_mock.ServiceMock{}
	controller := spendings.DefaultController(&repository, &pushService)
	_, err := controller.GetExpense(spendings.ExpenseId(uuid.New().String()), spendings.CounterpartyId(uuid.New().String()))
	if err == nil {
		t.Fatalf("`GetExpense` should be failed, found nil err")
	}
	if err.Code != spendings.GetDealErrorNotYourDeal {
		t.Fatalf("`GetExpense` should be failed with `not your expense`, found err %v", err)
	}
}

func TestGetExpenseOk(t *testing.T) {
	getCalls := 0
	actor := spendings.CounterpartyId(uuid.New().String())
	counterparty := spendings.CounterpartyId(uuid.New().String())
	repository := spendings_mock.RepositoryMock{
		GetExpenseImpl: func(id spendingsRepository.ExpenseId) (*spendingsRepository.IdentifiableExpense, error) {
			getCalls += 1
			return &spendingsRepository.IdentifiableExpense{
				Id: id,
				Expense: spendingsRepository.Expense{
					Shares: []spendingsRepository.ShareOfExpense{
						{
							Counterparty: spendingsRepository.CounterpartyId(actor),
						},
						{
							Counterparty: spendingsRepository.CounterpartyId(counterparty),
						},
					},
				},
			}, nil
		},
	}
	pushService := pushNotifications_mock.ServiceMock{}
	controller := spendings.DefaultController(&repository, &pushService)
	_, err := controller.GetExpense(spendings.ExpenseId(uuid.New().String()), actor)
	if err != nil {
		t.Fatalf("`GetExpense` should not be failed, found err %v", err)
	}
	if getCalls != 1 {
		t.Fatalf("get should be called once, found %d", getCalls)
	}
}

func TestGetExpensesWithFailedToGetFromRepository(t *testing.T) {
	repository := spendings_mock.RepositoryMock{
		GetExpensesBetweenImpl: func(counterparty1, counterparty2 spendingsRepository.CounterpartyId) ([]spendingsRepository.IdentifiableExpense, error) {
			return []spendingsRepository.IdentifiableExpense{}, errors.New("some error")
		},
	}
	pushService := pushNotifications_mock.ServiceMock{}
	controller := spendings.DefaultController(&repository, &pushService)
	_, err := controller.GetExpensesWith(spendings.CounterpartyId(uuid.New().String()), spendings.CounterpartyId(uuid.New().String()))
	if err == nil {
		t.Fatalf("`GetExpensesWith` should be failed, found nil err")
	}
	if err.Code != spendings.GetDealsErrorInternal {
		t.Fatalf("`GetExpensesWith` should be failed with `internal`, found err %v", err)
	}
}

func TestGetExpensesOk(t *testing.T) {
	getCalls := 0
	repository := spendings_mock.RepositoryMock{
		GetExpensesBetweenImpl: func(counterparty1, counterparty2 spendingsRepository.CounterpartyId) ([]spendingsRepository.IdentifiableExpense, error) {
			getCalls += 1
			return []spendingsRepository.IdentifiableExpense{}, nil
		},
	}
	pushService := pushNotifications_mock.ServiceMock{}
	controller := spendings.DefaultController(&repository, &pushService)
	_, err := controller.GetExpensesWith(spendings.CounterpartyId(uuid.New().String()), spendings.CounterpartyId(uuid.New().String()))
	if err != nil {
		t.Fatalf("`GetExpensesWith` should not be failed, found err %v", err)
	}
	if getCalls != 1 {
		t.Fatalf("get should be called once, found %d", getCalls)
	}
}

func TestGetBalanceWithFailedToGetFromRepository(t *testing.T) {
	repository := spendings_mock.RepositoryMock{
		GetBalanceImpl: func(counterparty spendingsRepository.CounterpartyId) ([]spendingsRepository.Balance, error) {
			return []spendingsRepository.Balance{}, errors.New("some error")
		},
	}
	pushService := pushNotifications_mock.ServiceMock{}
	controller := spendings.DefaultController(&repository, &pushService)
	_, err := controller.GetBalance(spendings.CounterpartyId(uuid.New().String()))
	if err == nil {
		t.Fatalf("`GetBalance` should be failed, found nil err")
	}
	if err.Code != spendings.GetCounterpartiesErrorInternal {
		t.Fatalf("`GetBalance` should be failed with `internal`, found err %v", err)
	}
}

func TestGetBalanceOk(t *testing.T) {
	getCalls := 0
	repository := spendings_mock.RepositoryMock{
		GetBalanceImpl: func(counterparty spendingsRepository.CounterpartyId) ([]spendingsRepository.Balance, error) {
			getCalls += 1
			return []spendingsRepository.Balance{}, nil
		},
	}
	pushService := pushNotifications_mock.ServiceMock{}
	controller := spendings.DefaultController(&repository, &pushService)
	_, err := controller.GetBalance(spendings.CounterpartyId(uuid.New().String()))
	if err != nil {
		t.Fatalf("`GetBalance` should not be failed, found err %v", err)
	}
	if getCalls != 1 {
		t.Fatalf("get should be called once, found %d", getCalls)
	}
}
