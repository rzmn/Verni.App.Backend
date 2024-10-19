package spendings

import (
	"verni/internal/common"
	"verni/internal/pushNotifications"
	"verni/internal/storage"
)

type UserId storage.UserId
type Deal storage.Deal
type DealId storage.DealId
type IdentifiableDeal storage.IdentifiableDeal
type SpendingsPreview storage.SpendingsPreview

type Controller interface {
	CreateDeal(deal Deal, userId UserId) *common.CodeBasedError[CreateDealErrorCode]
	DeleteDeal(dealId DealId, userId UserId) (IdentifiableDeal, *common.CodeBasedError[DeleteDealErrorCode])
	GetDeal(dealId DealId, userId UserId) (IdentifiableDeal, *common.CodeBasedError[GetDealErrorCode])
	GetDeals(counterparty UserId, userId UserId) ([]IdentifiableDeal, *common.CodeBasedError[GetDealsErrorCode])
	GetCounterparties(userId UserId) ([]SpendingsPreview, *common.CodeBasedError[GetCounterpartiesErrorCode])
}

func DefaultController(storage storage.Storage, pushNotifications pushNotifications.Service) Controller {
	return &defaultController{
		storage:           storage,
		pushNotifications: pushNotifications,
	}
}
