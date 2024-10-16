package spendings

import (
	"net/http"
	"verni/internal/apns"
	"verni/internal/auth/jwt"
	spendingsController "verni/internal/controllers/spendings"
	httpserver "verni/internal/http-server"
	"verni/internal/http-server/longpoll"
	"verni/internal/http-server/middleware"
	"verni/internal/http-server/responses"
	"verni/internal/storage"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine, db storage.Storage, jwtService jwt.Service, apns apns.Service, longpoll longpoll.Service) {
	ensureLoggedIn := middleware.EnsureLoggedIn(db, jwtService)
	hostFromToken := func(c *gin.Context) spendingsController.UserId {
		return spendingsController.UserId(c.Request.Header.Get(middleware.LoggedInSubjectKey))
	}
	controller := spendingsController.DefaultController(db)
	methodGroup := router.Group("/spendings", ensureLoggedIn)
	methodGroup.POST("/createDeal", func(c *gin.Context) {
		type CreateDealRequest struct {
			Deal storage.Deal `json:"deal"`
		}
		var request CreateDealRequest
		if err := c.BindJSON(&request); err != nil {
			httpserver.AnswerWithBadRequest(c, err)
			return
		}
		if err := controller.CreateDeal(spendingsController.Deal(request.Deal), hostFromToken(c)); err != nil {
			switch err.Code {
			default:
				httpserver.AnswerWithUnknownError(c, err)
			}
			return
		}
		c.JSON(http.StatusOK, responses.OK())
	})
	methodGroup.POST("/deleteDeal", func(c *gin.Context) {
		type DeleteDealRequest struct {
			DealId storage.DealId `json:"dealId"`
		}
		var request DeleteDealRequest
		if err := c.BindJSON(&request); err != nil {
			httpserver.AnswerWithBadRequest(c, err)
			return
		}
		_, err := controller.DeleteDeal(spendingsController.DealId(request.DealId), hostFromToken(c))
		if err != nil {
			switch err.Code {
			default:
				httpserver.AnswerWithUnknownError(c, err)
			}
			return
		}
		c.JSON(http.StatusOK, responses.OK())
	})
	methodGroup.GET("/getCounterparties", func(c *gin.Context) {
		preview, err := controller.GetCounterparties(hostFromToken(c))
		if err != nil {
			switch err.Code {
			default:
				httpserver.AnswerWithUnknownError(c, err)
			}
			return
		}
		c.JSON(http.StatusOK, responses.Success(preview))
	})
	methodGroup.GET("/getDeals", func(c *gin.Context) {
		type GetDealsRequest struct {
			Counterparty storage.UserId `json:"counterparty"`
		}
		var request GetDealsRequest
		if err := c.BindJSON(&request); err != nil {
			httpserver.AnswerWithBadRequest(c, err)
			return
		}
		deals, err := controller.GetDeals(spendingsController.UserId(request.Counterparty), hostFromToken(c))
		if err != nil {
			switch err.Code {
			default:
				httpserver.AnswerWithUnknownError(c, err)
			}
			return
		}
		c.JSON(http.StatusOK, responses.Success(deals))
	})
	methodGroup.GET("/getDeal", func(c *gin.Context) {
		type GetDealRequest struct {
			Id storage.DealId `json:"dealId"`
		}
		var request GetDealRequest
		if err := c.BindJSON(&request); err != nil {
			httpserver.AnswerWithBadRequest(c, err)
			return
		}
		deal, err := controller.GetDeal(spendingsController.DealId(request.Id), hostFromToken(c))
		if err != nil {
			switch err.Code {
			default:
				httpserver.AnswerWithUnknownError(c, err)
			}
			return
		}
		c.JSON(http.StatusOK, responses.Success(deal))
	})
}
