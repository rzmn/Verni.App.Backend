package users

import (
	"net/http"
	"verni/internal/auth/jwt"
	"verni/internal/common"
	httpserver "verni/internal/http-server"
	"verni/internal/http-server/middleware"
	"verni/internal/http-server/responses"
	usersController "verni/internal/http-server/router/users"
	"verni/internal/storage"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine, db storage.Storage, jwtService jwt.Service) {
	ensureLoggedIn := middleware.EnsureLoggedIn(db, jwtService)
	hostFromToken := func(c *gin.Context) usersController.UserId {
		return usersController.UserId(c.Request.Header.Get(middleware.LoggedInSubjectKey))
	}
	controller := usersController.DefaultController(db)
	methodGroup := router.Group("/users", ensureLoggedIn)
	methodGroup.GET("/get", func(c *gin.Context) {
		type GetUsersRequest struct {
			Ids []storage.UserId `json:"ids"`
		}
		var request GetUsersRequest
		if err := c.BindJSON(&request); err != nil {
			httpserver.AnswerWithBadRequest(c, err)
			return
		}
		users, err := controller.Get(common.Map(request.Ids, func(id storage.UserId) usersController.UserId {
			return usersController.UserId(id)
		}), hostFromToken(c))
		if err != nil {
			switch err.Code {
			default:
				httpserver.AnswerWithUnknownError(c, err)
			}
			return
		}
		c.JSON(http.StatusOK, responses.Success(users))
	})
	methodGroup.GET("/search", func(c *gin.Context) {
		type SearchUsersRequest struct {
			Query string `json:"query"`
		}
		var request SearchUsersRequest
		if err := c.BindJSON(&request); err != nil {
			httpserver.AnswerWithBadRequest(c, err)
			return
		}
		users, err := controller.Search(request.Query, hostFromToken(c))
		if err != nil {
			switch err.Code {
			default:
				httpserver.AnswerWithUnknownError(c, err)
			}
			return
		}
		c.JSON(http.StatusOK, responses.Success(users))
	})
}
