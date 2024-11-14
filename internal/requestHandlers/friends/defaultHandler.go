package friends

import (
	"net/http"
	"verni/internal/common"
	friendsController "verni/internal/controllers/friends"
	"verni/internal/schema"
	"verni/internal/services/logging"
	"verni/internal/services/longpoll"
	"verni/internal/services/pushNotifications"
)

type defaultRequestsHandler struct {
	controller     friendsController.Controller
	pushService    pushNotifications.Service
	pollingService longpoll.Service
	logger         logging.Service
}

func (c *defaultRequestsHandler) AcceptRequest(
	subject schema.UserId,
	request AcceptFriendRequest,
	success func(schema.StatusCode, schema.VoidResponse),
	failure func(schema.StatusCode, schema.Response[schema.Error]),
) {
	if err := c.controller.AcceptFriendRequest(friendsController.UserId(request.Sender), friendsController.UserId(subject)); err != nil {
		switch err.Code {
		case friendsController.AcceptFriendRequestErrorNoSuchRequest:
			failure(http.StatusConflict, schema.Failure(err, schema.CodeNoSuchRequest))
		default:
			c.logger.LogError("acceptRequest request %v failed with unknown err: %v", request, err)
			failure(http.StatusInternalServerError, schema.Failure(err, schema.CodeInternal))
		}
		return
	}
	success(http.StatusOK, schema.OK())
}

func (c *defaultRequestsHandler) GetFriends(
	subject schema.UserId,
	request GetFriendsRequest,
	success func(schema.StatusCode, schema.Response[map[schema.FriendStatus][]schema.UserId]),
	failure func(schema.StatusCode, schema.Response[schema.Error]),
) {
	friends, err := c.controller.GetFriends(common.Map(request.Statuses, func(status schema.FriendStatus) friendsController.FriendStatus {
		return friendsController.FriendStatus(status)
	}), friendsController.UserId(subject))
	if err != nil {
		switch err.Code {
		default:
			c.logger.LogError("getFriends request %v failed with unknown err: %v", request, err)
			failure(http.StatusInternalServerError, schema.Failure(err, schema.CodeInternal))
		}
		return
	}
	response := map[schema.FriendStatus][]schema.UserId{}
	for status, friends := range friends {
		response[schema.FriendStatus(status)] = common.Map(friends, func(id friendsController.UserId) schema.UserId {
			return schema.UserId(id)
		})
	}
	success(http.StatusOK, schema.Success(response))
}

func (c *defaultRequestsHandler) RejectRequest(
	subject schema.UserId,
	request RejectFriendRequest,
	success func(schema.StatusCode, schema.VoidResponse),
	failure func(schema.StatusCode, schema.Response[schema.Error]),
) {
	if err := c.controller.RollbackFriendRequest(friendsController.UserId(request.Sender), friendsController.UserId(subject)); err != nil {
		switch err.Code {
		case friendsController.RollbackFriendRequestErrorNoSuchRequest:
			failure(http.StatusConflict, schema.Failure(err, schema.CodeNoSuchRequest))
		default:
			c.logger.LogError("rejectRequest request %v failed with unknown err: %v", request, err)
			failure(http.StatusInternalServerError, schema.Failure(err, schema.CodeInternal))
		}
		return
	}
	success(http.StatusOK, schema.OK())
}

func (c *defaultRequestsHandler) RollbackRequest(
	subject schema.UserId,
	request RollbackFriendRequest,
	success func(schema.StatusCode, schema.VoidResponse),
	failure func(schema.StatusCode, schema.Response[schema.Error]),
) {
	if err := c.controller.RollbackFriendRequest(friendsController.UserId(subject), friendsController.UserId(request.Target)); err != nil {
		switch err.Code {
		case friendsController.RollbackFriendRequestErrorNoSuchRequest:
			failure(http.StatusConflict, schema.Failure(err, schema.CodeNoSuchRequest))
		default:
			c.logger.LogError("rollbackRequest request %v failed with unknown err: %v", request, err)
			failure(http.StatusInternalServerError, schema.Failure(err, schema.CodeInternal))
		}
		return
	}
	success(http.StatusOK, schema.OK())
}

func (c *defaultRequestsHandler) SendRequest(
	subject schema.UserId,
	request SendFriendRequest,
	success func(schema.StatusCode, schema.VoidResponse),
	failure func(schema.StatusCode, schema.Response[schema.Error]),
) {
	if err := c.controller.SendFriendRequest(friendsController.UserId(subject), friendsController.UserId(request.Target)); err != nil {
		switch err.Code {
		case friendsController.SendFriendRequestErrorAlreadySent:
			failure(http.StatusConflict, schema.Failure(err, schema.CodeAlreadySend))
		case friendsController.SendFriendRequestErrorHaveIncomingRequest:
			failure(http.StatusConflict, schema.Failure(err, schema.CodeHaveIncomingRequest))
		case friendsController.SendFriendRequestErrorAlreadyFriends:
			failure(http.StatusConflict, schema.Failure(err, schema.CodeAlreadyFriends))
		default:
			c.logger.LogError("sendRequest request %v failed with unknown err: %v", request, err)
			failure(http.StatusInternalServerError, schema.Failure(err, schema.CodeInternal))
		}
		return
	}
	success(http.StatusOK, schema.OK())
}

func (c *defaultRequestsHandler) Unfriend(
	subject schema.UserId,
	request UnfriendRequest,
	success func(schema.StatusCode, schema.VoidResponse),
	failure func(schema.StatusCode, schema.Response[schema.Error]),
) {
	if err := c.controller.Unfriend(friendsController.UserId(subject), friendsController.UserId(request.Target)); err != nil {
		switch err.Code {
		case friendsController.UnfriendErrorNotAFriend:
			failure(http.StatusConflict, schema.Failure(err, schema.CodeNotAFriend))
		default:
			c.logger.LogError("rollbackRequest request %v failed with unknown err: %v", request, err)
			failure(http.StatusInternalServerError, schema.Failure(err, schema.CodeInternal))
		}
		return
	}
	success(http.StatusOK, schema.OK())
}
