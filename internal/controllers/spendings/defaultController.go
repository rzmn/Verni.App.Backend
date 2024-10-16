package spendings

import (
	"log"
	"verni/internal/apns"
	"verni/internal/common"
	"verni/internal/storage"
)

type defaultController struct {
	storage storage.Storage
	apns    apns.Service
}

func (s *defaultController) CreateDeal(deal Deal, userId UserId) *common.CodeBasedError[CreateDealErrorCode] {
	const op = "spendings.defaultController.CreateDeal"
	log.Printf("%s: start[uid=%s]", op, userId)
	for i := 0; i < len(deal.Spendings); i++ {
		spending := deal.Spendings[i]
		if spending.UserId == storage.UserId(userId) {
			continue
		}
		exists, err := s.storage.IsUserExists(spending.UserId)
		if err != nil {
			log.Printf("%s: cannot check user existance in db err: %v", op, err)
			return common.NewErrorWithDescription(CreateDealErrorInternal, err.Error())
		}
		if !exists {
			log.Printf("%s: user %s does not exists", op, spending.UserId)
			return common.NewError(CreateDealErrorNoSuchUser)
		}
	}
	dealId, err := s.storage.InsertDeal(storage.Deal(deal))
	if err != nil {
		log.Printf("%s: cannot insert deal into db err: %v", op, err)
		return common.NewErrorWithDescription(CreateDealErrorInternal, err.Error())
	}
	for i := 0; i < len(deal.Spendings); i++ {
		spending := deal.Spendings[i]
		if spending.UserId == storage.UserId(userId) {
			continue
		}
		s.apns.NewExpenseReceived(apns.UserId(spending.UserId), apns.Deal{
			Deal: storage.Deal(deal),
			Id:   dealId,
		}, apns.UserId(userId))
	}
	log.Printf("%s: success[uid=%s]", op, userId)
	return nil
}

func (s *defaultController) DeleteDeal(dealId DealId, userId UserId) (IdentifiableDeal, *common.CodeBasedError[DeleteDealErrorCode]) {
	const op = "spendings.defaultController.DeleteDeal"
	log.Printf("%s: start[did=%s uid=%s]", op, dealId, userId)
	dealFromDb, err := s.storage.GetDeal(storage.DealId(dealId))
	if err != nil {
		log.Printf("%s: cannot get deal from db err: %v", op, err)
		return IdentifiableDeal{}, common.NewErrorWithDescription(DeleteDealErrorInternal, err.Error())
	}
	if dealFromDb == nil {
		log.Printf("%s: deal %s does not exists", op, dealId)
		return IdentifiableDeal{}, common.NewError(DeleteDealErrorDealNotFound)
	}
	counterparties, err := s.storage.GetCounterpartiesForDeal(storage.DealId(dealId))
	if err != nil {
		log.Printf("%s: cannot get deal counterparties from db err: %v", op, err)
		return IdentifiableDeal{}, common.NewErrorWithDescription(DeleteDealErrorInternal, err.Error())
	}
	var isYourDeal bool
	for i := 0; i < len(counterparties); i++ {
		if counterparties[i] == storage.UserId(userId) {
			isYourDeal = true
			break
		}
	}
	if !isYourDeal {
		log.Printf("%s: user %s is not found in deals %s counterparties", op, userId, dealId)
		return IdentifiableDeal{}, common.NewError(DeleteDealErrorNotYourDeal)
	}
	if err := s.storage.RemoveDeal(storage.DealId(dealId)); err != nil {
		log.Printf("%s: cannot remove deal from db err: %v", op, err)
		return IdentifiableDeal{}, common.NewErrorWithDescription(DeleteDealErrorInternal, err.Error())
	}
	log.Printf("%s: success[did=%s uid=%s]", op, dealId, userId)
	return IdentifiableDeal(*dealFromDb), nil
}

func (s *defaultController) GetDeal(dealId DealId, userId UserId) (IdentifiableDeal, *common.CodeBasedError[GetDealErrorCode]) {
	const op = "spendings.defaultController.GetDeal"
	log.Printf("%s: start[did=%s uid=%s]", op, dealId, userId)
	deal, err := s.storage.GetDeal(storage.DealId(dealId))
	if err != nil {
		log.Printf("%s: cannot get deal from db err: %v", op, err)
		return IdentifiableDeal{}, common.NewErrorWithDescription(GetDealErrorInternal, err.Error())
	}
	if deal == nil {
		log.Printf("%s: deal %s is not found in db", op, dealId)
		return IdentifiableDeal{}, common.NewError(GetDealErrorDealNotFound)
	}
	log.Printf("%s: success[did=%s uid=%s]", op, dealId, userId)
	return IdentifiableDeal(*deal), nil
}

func (s *defaultController) GetDeals(counterparty UserId, userId UserId) ([]IdentifiableDeal, *common.CodeBasedError[GetDealsErrorCode]) {
	const op = "spendings.defaultController.GetDeals"
	log.Printf("%s: start[counterparty=%s uid=%s]", op, counterparty, userId)
	exists, err := s.storage.IsUserExists(storage.UserId(counterparty))
	if err != nil {
		log.Printf("%s: cannot check counterparty %s exists in db err: %v", op, counterparty, err)
		return []IdentifiableDeal{}, common.NewErrorWithDescription(GetDealsErrorInternal, err.Error())
	}
	if !exists {
		log.Printf("%s: counterparty %s does not found", op, userId)
		return []IdentifiableDeal{}, common.NewError(GetDealsErrorNoSuchUser)
	}
	deals, err := s.storage.GetDeals(storage.UserId(counterparty), storage.UserId(userId))
	if err != nil {
		log.Printf("%s: cannot get deals from db err: %v", op, err)
		return []IdentifiableDeal{}, common.NewErrorWithDescription(GetDealsErrorInternal, err.Error())
	}
	log.Printf("%s: success[counterparty=%s uid=%s]", op, counterparty, userId)
	return common.Map(deals, func(deal storage.IdentifiableDeal) IdentifiableDeal {
		return IdentifiableDeal(deal)
	}), nil
}

func (s *defaultController) GetCounterparties(userId UserId) ([]SpendingsPreview, *common.CodeBasedError[GetCounterpartiesErrorCode]) {
	const op = "spendings.defaultController.GetCounterparties"
	log.Printf("%s: start[uid=%s]", op, userId)
	preview, err := s.storage.GetCounterparties(storage.UserId(userId))
	if err != nil {
		log.Printf("%s: cannot get spendings preview for %s from db err: %v", op, userId, err)
		return []SpendingsPreview{}, common.NewErrorWithDescription(GetCounterpartiesErrorInternal, err.Error())
	}
	log.Printf("%s: success[uid=%s]", op, userId)
	return common.Map(preview, func(preview storage.SpendingsPreview) SpendingsPreview {
		return SpendingsPreview(preview)
	}), nil
}
