package users

import (
	"net/http"
	"verni/internal/common"
	usersController "verni/internal/controllers/users"
	httpserver "verni/internal/http-server"
	"verni/internal/http-server/middleware"
	"verni/internal/http-server/responses"

	"github.com/gin-gonic/gin"
)

type UsersController usersController.Controller

func RegisterRoutes(
	router *gin.Engine,
	tokenChecker middleware.AccessTokenChecker,
	users UsersController,
) {
	methodGroup := router.Group("/users", tokenChecker.Handler)
	methodGroup.GET("/get", func(c *gin.Context) {
		type GetUsersRequest struct {
			Ids []httpserver.UserId `json:"ids"`
		}
		var request GetUsersRequest
		if err := c.BindJSON(&request); err != nil {
			httpserver.AnswerWithBadRequest(c, err)
			return
		}
		users, err := users.Get(common.Map(request.Ids, func(id httpserver.UserId) usersController.UserId {
			return usersController.UserId(id)
		}), usersController.UserId(tokenChecker.AccessToken(c)))
		if err != nil {
			switch err.Code {
			default:
				httpserver.AnswerWithUnknownError(c, err)
			}
			return
		}
		c.JSON(http.StatusOK, responses.Success(common.Map(users, mapUser)))
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
		users, err := users.Search(request.Query, usersController.UserId(tokenChecker.AccessToken(c)))
		if err != nil {
			switch err.Code {
			default:
				httpserver.AnswerWithUnknownError(c, err)
			}
			return
		}
		c.JSON(http.StatusOK, responses.Success(responses.Success(common.Map(users, mapUser))))
	})
}

func mapUser(user usersController.User) httpserver.User {
	return httpserver.User{
		Id:           httpserver.UserId(user.Id),
		DisplayName:  user.DisplayName,
		AvatarId:     (*httpserver.ImageId)(user.AvatarId),
		FriendStatus: httpserver.FriendStatus(user.FriendStatus),
	}
}
