package spendings

import (
	"database/sql"
	"fmt"
	"log"
	"verni/internal/repositories"

	_ "github.com/lib/pq"
)

type ExpenseId string
type CounterpartyId string
type Currency string
type Cost int64

type ShareOfExpense struct {
	Counterparty CounterpartyId
	Cost         Cost
}

type Expense struct {
	Timestamp int64
	Details   string
	Total     Cost
	Currency  Currency
	Shares    []ShareOfExpense
}

type IdentifiableExpense struct {
	Expense
	Id ExpenseId
}

type Balance struct {
	Counterparty CounterpartyId
	Currencies   map[Currency]Cost
}

type Repository interface {
	AddExpense(id Expense) repositories.MutationWorkItemWithReturnValue[ExpenseId]
	RemoveExpense(id ExpenseId) repositories.MutationWorkItem

	GetExpense(id ExpenseId) (*IdentifiableExpense, error)

	GetExpensesBetween(counterparty1 CounterpartyId, counterparty2 CounterpartyId) ([]IdentifiableExpense, error)
	GetBalance(counterparty CounterpartyId) ([]Balance, error)
}

func PostgresRepository(config repositories.PostgresConfig) (Repository, error) {
	const op = "repositories.friends.PostgresRepository"
	psqlConnection := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.Host,
		config.Port,
		config.User,
		config.Password,
		config.DbName,
	)
	db, err := sql.Open("postgres", psqlConnection)
	if err != nil {
		log.Printf("%s: open db failed err: %v", op, err)
		return &postgresRepository{}, err
	}
	return &postgresRepository{
		db: db,
	}, nil
}
