package friends

import (
	"net/http"
	"verni/internal/common"
	friendsController "verni/internal/controllers/friends"
	httpserver "verni/internal/http-server"
	"verni/internal/http-server/longpoll"
	"verni/internal/http-server/responses"
	"verni/internal/services/logging"
	"verni/internal/services/pushNotifications"
)

type defaultRequestsHandler struct {
	controller     friendsController.Controller
	pushService    pushNotifications.Service
	pollingService longpoll.Service
	logger         logging.Service
}

func (c *defaultRequestsHandler) AcceptRequest(
	subject friendsController.UserId,
	request AcceptFriendRequest,
	success func(HttpCode, responses.VoidResponse),
	failure func(HttpCode, responses.Response[responses.Error]),
) {
	if err := c.controller.AcceptFriendRequest(friendsController.UserId(request.Sender), subject); err != nil {
		switch err.Code {
		case friendsController.AcceptFriendRequestErrorNoSuchRequest:
			failure(
				http.StatusConflict,
				responses.Failure(
					common.NewErrorWithDescriptionValue(
						responses.CodeNoSuchRequest,
						err.Error(),
					),
				),
			)
		default:
			c.logger.LogError("acceptRequest request %v failed with unknown err: %v", request, err)
			failure(
				http.StatusInternalServerError,
				responses.Failure(
					common.NewErrorWithDescriptionValue(
						responses.CodeInternal,
						err.Error(),
					),
				),
			)
		}
		return
	}
	success(http.StatusOK, responses.OK())
}

func (c *defaultRequestsHandler) GetFriends(
	subject friendsController.UserId,
	request GetFriendsRequest,
	success func(HttpCode, responses.Response[map[httpserver.FriendStatus][]httpserver.UserId]),
	failure func(HttpCode, responses.Response[responses.Error]),
) {
	friends, err := c.controller.GetFriends(common.Map(request.Statuses, func(status httpserver.FriendStatus) friendsController.FriendStatus {
		return friendsController.FriendStatus(status)
	}), subject)
	if err != nil {
		switch err.Code {
		default:
			c.logger.LogError("getFriends request %v failed with unknown err: %v", request, err)
			failure(
				http.StatusInternalServerError,
				responses.Failure(
					common.NewErrorWithDescriptionValue(
						responses.CodeInternal,
						err.Error(),
					),
				),
			)
		}
		return
	}
	response := map[httpserver.FriendStatus][]httpserver.UserId{}
	for status, friends := range friends {
		response[httpserver.FriendStatus(status)] = common.Map(friends, func(id friendsController.UserId) httpserver.UserId {
			return httpserver.UserId(id)
		})
	}
	success(http.StatusOK, responses.Success(response))
}

func (c *defaultRequestsHandler) RejectRequest(
	subject friendsController.UserId,
	request RejectFriendRequest,
	success func(HttpCode, responses.VoidResponse),
	failure func(HttpCode, responses.Response[responses.Error]),
) {
	if err := c.controller.RollbackFriendRequest(friendsController.UserId(request.Sender), subject); err != nil {
		switch err.Code {
		case friendsController.RollbackFriendRequestErrorNoSuchRequest:
			failure(
				http.StatusConflict,
				responses.Failure(
					common.NewErrorWithDescriptionValue(
						responses.CodeNoSuchRequest,
						err.Error(),
					),
				),
			)
		default:
			c.logger.LogError("rejectRequest request %v failed with unknown err: %v", request, err)
			failure(
				http.StatusInternalServerError,
				responses.Failure(
					common.NewErrorWithDescriptionValue(
						responses.CodeInternal,
						err.Error(),
					),
				),
			)
		}
		return
	}
	success(http.StatusOK, responses.OK())
}

func (c *defaultRequestsHandler) RollbackRequest(
	subject friendsController.UserId,
	request RollbackFriendRequest,
	success func(HttpCode, responses.VoidResponse),
	failure func(HttpCode, responses.Response[responses.Error]),
) {
	if err := c.controller.RollbackFriendRequest(subject, friendsController.UserId(request.Target)); err != nil {
		switch err.Code {
		case friendsController.RollbackFriendRequestErrorNoSuchRequest:
			failure(
				http.StatusConflict,
				responses.Failure(
					common.NewErrorWithDescriptionValue(
						responses.CodeNoSuchRequest,
						err.Error(),
					),
				),
			)
		default:
			c.logger.LogError("rollbackRequest request %v failed with unknown err: %v", request, err)
			failure(
				http.StatusInternalServerError,
				responses.Failure(
					common.NewErrorWithDescriptionValue(
						responses.CodeInternal,
						err.Error(),
					),
				),
			)
		}
		return
	}
	success(http.StatusOK, responses.OK())
}

func (c *defaultRequestsHandler) SendRequest(
	subject friendsController.UserId,
	request SendFriendRequest,
	success func(HttpCode, responses.VoidResponse),
	failure func(HttpCode, responses.Response[responses.Error]),
) {
	if err := c.controller.SendFriendRequest(subject, friendsController.UserId(request.Target)); err != nil {
		switch err.Code {
		case friendsController.SendFriendRequestErrorAlreadySent:
			failure(
				http.StatusConflict,
				responses.Failure(
					common.NewErrorWithDescriptionValue(
						responses.CodeAlreadySend,
						err.Error(),
					),
				),
			)
		case friendsController.SendFriendRequestErrorHaveIncomingRequest:
			failure(
				http.StatusConflict,
				responses.Failure(
					common.NewErrorWithDescriptionValue(
						responses.CodeHaveIncomingRequest,
						err.Error(),
					),
				),
			)
		case friendsController.SendFriendRequestErrorAlreadyFriends:
			failure(
				http.StatusConflict,
				responses.Failure(
					common.NewErrorWithDescriptionValue(
						responses.CodeAlreadyFriends,
						err.Error(),
					),
				),
			)
		default:
			c.logger.LogError("sendRequest request %v failed with unknown err: %v", request, err)
			failure(
				http.StatusInternalServerError,
				responses.Failure(
					common.NewErrorWithDescriptionValue(
						responses.CodeInternal,
						err.Error(),
					),
				),
			)
		}
		return
	}
	success(http.StatusOK, responses.OK())
}

func (c *defaultRequestsHandler) Unfriend(
	subject friendsController.UserId,
	request UnfriendRequest,
	success func(HttpCode, responses.VoidResponse),
	failure func(HttpCode, responses.Response[responses.Error]),
) {
	if err := c.controller.Unfriend(subject, friendsController.UserId(request.Target)); err != nil {
		switch err.Code {
		case friendsController.UnfriendErrorNotAFriend:
			failure(
				http.StatusConflict,
				responses.Failure(
					common.NewErrorWithDescriptionValue(
						responses.CodeNotAFriend,
						err.Error(),
					),
				),
			)
		default:
			c.logger.LogError("rollbackRequest request %v failed with unknown err: %v", request, err)
			failure(
				http.StatusInternalServerError,
				responses.Failure(
					common.NewErrorWithDescriptionValue(
						responses.CodeInternal,
						err.Error(),
					),
				),
			)
		}
		return
	}
	success(http.StatusOK, responses.OK())
}
