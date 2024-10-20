package spendings

import (
	"database/sql"
	"fmt"
	"log"
	"verni/internal/repositories"
	"verni/internal/storage"

	_ "github.com/lib/pq"
)

type Deal storage.Deal
type DealId storage.DealId
type IdentifiableDeal storage.IdentifiableDeal
type UserId storage.UserId
type SpendingsPreview storage.SpendingsPreview

type Repository interface {
	InsertDeal(deal Deal) repositories.MutationWorkItemWithReturnValue[DealId]
	RemoveDeal(did DealId) repositories.MutationWorkItem

	GetDeal(did DealId) (*IdentifiableDeal, error)

	GetDeals(counterparty1 UserId, counterparty2 UserId) ([]IdentifiableDeal, error)
	GetCounterparties(uid UserId) ([]SpendingsPreview, error)
	GetCounterpartiesForDeal(did DealId) ([]UserId, error)
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
