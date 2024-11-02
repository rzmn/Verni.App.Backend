package spendings

import (
	"net/http"
	"verni/internal/common"
	spendingsController "verni/internal/controllers/spendings"
	httpserver "verni/internal/http-server"
	"verni/internal/http-server/middleware"
	"verni/internal/http-server/responses"
	spendingsRepository "verni/internal/repositories/spendings"

	"github.com/gin-gonic/gin"
)

type SpendingController spendingsController.Controller

func RegisterRoutes(
	router *gin.Engine,
	tokenChecker middleware.AccessTokenChecker,
	spendings SpendingController,
) {
	methodGroup := router.Group("/spendings", tokenChecker.Handler)
	methodGroup.POST("/addExpense", func(c *gin.Context) {
		type AddExpenseRequest struct {
			Expense httpserver.Expense `json:"expense"`
		}
		var request AddExpenseRequest
		if err := c.BindJSON(&request); err != nil {
			httpserver.AnswerWithBadRequest(c, err)
			return
		}
		expense := spendingsController.Expense{
			Timestamp: request.Expense.Timestamp,
			Details:   request.Expense.Details,
			Total:     spendingsRepository.Cost(request.Expense.Total),
			Currency:  spendingsRepository.Currency(request.Expense.Currency),
			Shares: common.Map(request.Expense.Shares, func(share httpserver.ShareOfExpense) spendingsRepository.ShareOfExpense {
				return spendingsRepository.ShareOfExpense{
					Counterparty: spendingsRepository.CounterpartyId(share.UserId),
					Cost:         spendingsRepository.Cost(share.Cost),
				}
			}),
		}
		if err := spendings.AddExpense(expense, spendingsController.CounterpartyId(tokenChecker.AccessToken(c))); err != nil {
			switch err.Code {
			case spendingsController.CreateDealErrorNoSuchUser:
				httpserver.Answer(c, err, http.StatusConflict, responses.CodeNoSuchUser)
			case spendingsController.CreateDealErrorNotYourExpense:
				httpserver.Answer(c, err, http.StatusConflict, responses.CodeIsNotYourDeal)
			default:
				httpserver.AnswerWithUnknownError(c, err)
			}
			return
		}
		c.JSON(http.StatusOK, responses.OK())
	})
	methodGroup.POST("/removeExpense", func(c *gin.Context) {
		type RemoveExpenseRequest struct {
			ExpenseId httpserver.ExpenseId `json:"expenseId"`
		}
		var request RemoveExpenseRequest
		if err := c.BindJSON(&request); err != nil {
			httpserver.AnswerWithBadRequest(c, err)
			return
		}
		_, err := spendings.RemoveExpense(spendingsController.ExpenseId(request.ExpenseId), spendingsController.CounterpartyId(tokenChecker.AccessToken(c)))
		if err != nil {
			switch err.Code {
			case spendingsController.DeleteDealErrorDealNotFound:
				httpserver.Answer(c, err, http.StatusConflict, responses.CodeDealNotFound)
			case spendingsController.DeleteDealErrorNotAFriend:
				httpserver.Answer(c, err, http.StatusConflict, responses.CodeNotAFriend)
			case spendingsController.DeleteDealErrorNotYourExpense:
				httpserver.Answer(c, err, http.StatusConflict, responses.CodeIsNotYourDeal)
			default:
				httpserver.AnswerWithUnknownError(c, err)
			}
			return
		}
		c.JSON(http.StatusOK, responses.OK())
	})
	methodGroup.GET("/getBalance", func(c *gin.Context) {
		balance, err := spendings.GetBalance(spendingsController.CounterpartyId(tokenChecker.AccessToken(c)))
		if err != nil {
			switch err.Code {
			default:
				httpserver.AnswerWithUnknownError(c, err)
			}
			return
		}
		response := common.Map(balance, mapBalance)
		c.JSON(http.StatusOK, responses.Success(response))
	})
	methodGroup.GET("/getExpenses", func(c *gin.Context) {
		type GetExpensesRequest struct {
			Counterparty httpserver.UserId `json:"counterparty"`
		}
		var request GetExpensesRequest
		if err := c.BindJSON(&request); err != nil {
			httpserver.AnswerWithBadRequest(c, err)
			return
		}
		expenses, err := spendings.GetExpensesWith(spendingsController.CounterpartyId(request.Counterparty), spendingsController.CounterpartyId(tokenChecker.AccessToken(c)))
		if err != nil {
			switch err.Code {
			default:
				httpserver.AnswerWithUnknownError(c, err)
			}
			return
		}
		response := common.Map(expenses, mapIdentifiableExpense)
		c.JSON(http.StatusOK, responses.Success(response))
	})
	methodGroup.GET("/getExpense", func(c *gin.Context) {
		type GetExpenseRequest struct {
			Id httpserver.ExpenseId `json:"id"`
		}
		var request GetExpenseRequest
		if err := c.BindJSON(&request); err != nil {
			httpserver.AnswerWithBadRequest(c, err)
			return
		}
		expense, err := spendings.GetExpense(spendingsController.ExpenseId(request.Id), spendingsController.CounterpartyId(tokenChecker.AccessToken(c)))
		if err != nil {
			switch err.Code {
			case spendingsController.GetDealErrorDealNotFound:
				httpserver.Answer(c, err, http.StatusConflict, responses.CodeDealNotFound)
			case spendingsController.GetDealErrorNotYourDeal:
				httpserver.Answer(c, err, http.StatusConflict, responses.CodeIsNotYourDeal)
			default:
				httpserver.AnswerWithUnknownError(c, err)
			}
			return
		}
		c.JSON(http.StatusOK, responses.Success(mapIdentifiableExpense(expense)))
	})
}

func mapIdentifiableExpense(expense spendingsController.IdentifiableExpense) httpserver.IdentifiableExpense {
	return httpserver.IdentifiableExpense{
		Id:      httpserver.ExpenseId(expense.Id),
		Expense: mapExpense(spendingsController.Expense(expense.Expense)),
	}
}

func mapExpense(expense spendingsController.Expense) httpserver.Expense {
	return httpserver.Expense{
		Timestamp:   expense.Timestamp,
		Details:     expense.Details,
		Total:       httpserver.Cost(expense.Total),
		Attachments: []httpserver.ExpenseAttachment{},
		Currency:    httpserver.Currency(expense.Currency),
		Shares:      common.Map(expense.Shares, mapShareOfExpense),
	}
}

func mapShareOfExpense(share spendingsRepository.ShareOfExpense) httpserver.ShareOfExpense {
	return httpserver.ShareOfExpense{
		UserId: httpserver.UserId(share.Counterparty),
		Cost:   httpserver.Cost(share.Cost),
	}
}

func mapBalance(balance spendingsController.Balance) httpserver.Balance {
	currencies := map[httpserver.Currency]httpserver.Cost{}
	for currency, cost := range balance.Currencies {
		currencies[httpserver.Currency(currency)] = httpserver.Cost(cost)
	}
	return httpserver.Balance{
		Counterparty: string(balance.Counterparty),
		Currencies:   currencies,
	}
}
