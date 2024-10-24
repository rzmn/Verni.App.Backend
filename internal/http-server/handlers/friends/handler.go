package friends

import (
	"net/http"
	friendsController "verni/internal/controllers/friends"
	httpserver "verni/internal/http-server"
	"verni/internal/http-server/middleware"
	"verni/internal/http-server/responses"

	"github.com/gin-gonic/gin"
)

type FriendsController friendsController.Controller

func RegisterRoutes(
	router *gin.Engine,
	tokenChecker middleware.AccessTokenChecker,
	friends FriendsController,
) {
	methodGroup := router.Group("/friends", tokenChecker.Handler)
	methodGroup.POST("/acceptRequest", func(c *gin.Context) {
		type AcceptFriendRequest struct {
			Sender httpserver.UserId `json:"sender"`
		}
		var request AcceptFriendRequest
		if err := c.BindJSON(&request); err != nil {
			httpserver.AnswerWithBadRequest(c, err)
			return
		}
		if err := friends.AcceptFriendRequest(friendsController.UserId(request.Sender), friendsController.UserId(tokenChecker.AccessToken(c))); err != nil {
			switch err.Code {
			case friendsController.AcceptFriendRequestErrorNoSuchRequest:
				httpserver.Answer(c, err, http.StatusConflict, responses.CodeNoSuchRequest)
			default:
				httpserver.AnswerWithUnknownError(c, err)
			}
			return
		}
		c.JSON(http.StatusOK, responses.OK())
	})
	methodGroup.GET("/get", func(c *gin.Context) {
		type GetFriendsRequest struct {
			Statuses []friendsController.FriendStatus `json:"statuses"`
		}
		var request GetFriendsRequest
		if err := c.BindJSON(&request); err != nil {
			httpserver.AnswerWithBadRequest(c, err)
			return
		}
		friends, err := friends.GetFriends(request.Statuses, friendsController.UserId(tokenChecker.AccessToken(c)))
		if err != nil {
			switch err.Code {
			default:
				httpserver.AnswerWithUnknownError(c, err)
			}
			return
		}
		c.JSON(http.StatusOK, responses.Success(friends))
	})
	methodGroup.POST("/rejectRequest", func(c *gin.Context) {
		type RejectFriendRequest struct {
			Sender httpserver.UserId `json:"sender"`
		}
		var request RejectFriendRequest
		if err := c.BindJSON(&request); err != nil {
			httpserver.AnswerWithBadRequest(c, err)
			return
		}
		if err := friends.RollbackFriendRequest(friendsController.UserId(tokenChecker.AccessToken(c)), friendsController.UserId(request.Sender)); err != nil {
			switch err.Code {
			case friendsController.RejectFriendRequestErrorNoSuchRequest:
				httpserver.Answer(c, err, http.StatusConflict, responses.CodeNoSuchRequest)
			default:
				httpserver.AnswerWithUnknownError(c, err)
			}
			return
		}
		c.JSON(http.StatusOK, responses.OK())
	})
	methodGroup.POST("/rollbackRequest", func(c *gin.Context) {
		type RollbackFriendRequest struct {
			Target httpserver.UserId `json:"target"`
		}
		var request RollbackFriendRequest
		if err := c.BindJSON(&request); err != nil {
			httpserver.AnswerWithBadRequest(c, err)
			return
		}
		if err := friends.RollbackFriendRequest(friendsController.UserId(tokenChecker.AccessToken(c)), friendsController.UserId(request.Target)); err != nil {
			switch err.Code {
			case friendsController.RejectFriendRequestErrorAlreadyFriends:
				httpserver.Answer(c, err, http.StatusConflict, responses.CodeAlreadyFriends)
			default:
				httpserver.AnswerWithUnknownError(c, err)
			}
			return
		}
		c.JSON(http.StatusOK, responses.OK())
	})
	methodGroup.POST("/sendRequest", func(c *gin.Context) {
		type SendFriendRequest struct {
			Target httpserver.UserId `json:"target"`
		}
		var request SendFriendRequest
		if err := c.BindJSON(&request); err != nil {
			httpserver.AnswerWithBadRequest(c, err)
			return
		}
		if err := friends.SendFriendRequest(friendsController.UserId(tokenChecker.AccessToken(c)), friendsController.UserId(request.Target)); err != nil {
			switch err.Code {
			case friendsController.SendFriendRequestErrorAlreadySent:
				httpserver.Answer(c, err, http.StatusConflict, responses.CodeAlreadySend)
			case friendsController.SendFriendRequestErrorHaveIncomingRequest:
				httpserver.Answer(c, err, http.StatusConflict, responses.CodeHaveIncomingRequest)
			case friendsController.SendFriendRequestErrorAlreadyFriends:
				httpserver.Answer(c, err, http.StatusConflict, responses.CodeAlreadyFriends)
			default:
				httpserver.AnswerWithUnknownError(c, err)
			}
			return
		}
		c.JSON(http.StatusOK, responses.OK())
	})
	methodGroup.POST("/unfriend", func(c *gin.Context) {
		type SendFriendRequest struct {
			Target httpserver.UserId `json:"target"`
		}
		var request SendFriendRequest
		if err := c.BindJSON(&request); err != nil {
			httpserver.AnswerWithBadRequest(c, err)
			return
		}
		if err := friends.Unfriend(friendsController.UserId(tokenChecker.AccessToken(c)), friendsController.UserId(request.Target)); err != nil {
			switch err.Code {
			case friendsController.UnfriendErrorNotAFriend:
				httpserver.Answer(c, err, http.StatusConflict, responses.CodeNotAFriend)
			default:
				httpserver.AnswerWithUnknownError(c, err)
			}
			return
		}
		c.JSON(http.StatusOK, responses.OK())
	})
}
