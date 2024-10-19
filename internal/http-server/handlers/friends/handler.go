package friends

import (
	"net/http"
	"verni/internal/auth/jwt"
	friendsController "verni/internal/controllers/friends"
	httpserver "verni/internal/http-server"
	"verni/internal/http-server/longpoll"
	"verni/internal/http-server/middleware"
	"verni/internal/http-server/responses"
	"verni/internal/pushNotifications"
	"verni/internal/storage"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine, db storage.Storage, jwtService jwt.Service, apns pushNotifications.Service, longpoll longpoll.Service) {
	ensureLoggedIn := middleware.EnsureLoggedIn(db, jwtService)
	hostFromToken := func(c *gin.Context) friendsController.UserId {
		return friendsController.UserId(c.Request.Header.Get(middleware.LoggedInSubjectKey))
	}
	controller := friendsController.DefaultController(db)
	methodGroup := router.Group("/friends", ensureLoggedIn)
	methodGroup.POST("/acceptRequest", func(c *gin.Context) {
		type AcceptFriendRequest struct {
			Sender storage.UserId `json:"sender"`
		}
		var request AcceptFriendRequest
		if err := c.BindJSON(&request); err != nil {
			httpserver.AnswerWithBadRequest(c, err)
			return
		}
		if err := controller.AcceptFriendRequest(friendsController.UserId(request.Sender), hostFromToken(c)); err != nil {
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
		friends, err := controller.GetFriends(request.Statuses, hostFromToken(c))
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
			Sender storage.UserId `json:"sender"`
		}
		var request RejectFriendRequest
		if err := c.BindJSON(&request); err != nil {
			httpserver.AnswerWithBadRequest(c, err)
			return
		}
		if err := controller.RejectFriendRequest(friendsController.UserId(request.Sender), hostFromToken(c)); err != nil {
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
			Target storage.UserId `json:"target"`
		}
		var request RollbackFriendRequest
		if err := c.BindJSON(&request); err != nil {
			httpserver.AnswerWithBadRequest(c, err)
			return
		}
		if err := controller.RollbackFriendRequest(hostFromToken(c), friendsController.UserId(request.Target)); err != nil {
			switch err.Code {
			case friendsController.RollbackFriendRequestErrorNoSuchRequest:
				httpserver.Answer(c, err, http.StatusConflict, responses.CodeNoSuchRequest)
			default:
				httpserver.AnswerWithUnknownError(c, err)
			}
			return
		}
		c.JSON(http.StatusOK, responses.OK())
	})
	methodGroup.POST("/sendRequest", func(c *gin.Context) {
		type SendFriendRequest struct {
			Target storage.UserId `json:"target"`
		}
		var request SendFriendRequest
		if err := c.BindJSON(&request); err != nil {
			httpserver.AnswerWithBadRequest(c, err)
			return
		}
		if err := controller.SendFriendRequest(hostFromToken(c), friendsController.UserId(request.Target)); err != nil {
			switch err.Code {
			case friendsController.SendFriendRequestErrorAlreadySent:
				httpserver.Answer(c, err, http.StatusConflict, responses.CodeAlreadySend)
			case friendsController.SendFriendRequestErrorHaveIncomingRequest:
				httpserver.Answer(c, err, http.StatusConflict, responses.CodeHaveIncomingRequest)
			case friendsController.SendFriendRequestErrorAlreadyFriends:
				httpserver.Answer(c, err, http.StatusConflict, responses.CodeAlreadyFriends)
			case friendsController.SendFriendRequestErrorNoSuchUser:
				httpserver.Answer(c, err, http.StatusConflict, responses.CodeNoSuchUser)
			default:
				httpserver.AnswerWithUnknownError(c, err)
			}
			return
		}
		c.JSON(http.StatusOK, responses.OK())
	})
	methodGroup.POST("/unfriend", func(c *gin.Context) {
		type SendFriendRequest struct {
			Target storage.UserId `json:"target"`
		}
		var request SendFriendRequest
		if err := c.BindJSON(&request); err != nil {
			httpserver.AnswerWithBadRequest(c, err)
			return
		}
		if err := controller.Unfriend(hostFromToken(c), friendsController.UserId(request.Target)); err != nil {
			switch err.Code {
			case friendsController.UnfriendErrorNoSuchUser:
				httpserver.Answer(c, err, http.StatusConflict, responses.CodeNoSuchUser)
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
