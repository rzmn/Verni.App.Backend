package defaultFriendsHandler

import (
	"net/http"

	"github.com/rzmn/governi/internal/common"
	friendsController "github.com/rzmn/governi/internal/controllers/friends"
	"github.com/rzmn/governi/internal/requestHandlers/friends"
	"github.com/rzmn/governi/internal/schema"
	"github.com/rzmn/governi/internal/services/logging"
	"github.com/rzmn/governi/internal/services/pushNotifications"
	"github.com/rzmn/governi/internal/services/realtimeEvents"
)

func New(
	controller friendsController.Controller,
	pushService pushNotifications.Service,
	realtimeEvents realtimeEvents.Service,
	logger logging.Service,
) friends.RequestsHandler {
	return &defaultRequestsHandler{
		controller:     controller,
		pushService:    pushService,
		realtimeEvents: realtimeEvents,
		logger:         logger,
	}
}

type defaultRequestsHandler struct {
	controller     friendsController.Controller
	pushService    pushNotifications.Service
	realtimeEvents realtimeEvents.Service
	logger         logging.Service
}

func (c *defaultRequestsHandler) AcceptRequest(
	subject schema.UserId,
	request schema.AcceptFriendRequest,
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
	request schema.GetFriendsRequest,
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
	request schema.RejectFriendRequest,
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
	request schema.RollbackFriendRequest,
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
	request schema.SendFriendRequest,
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
	request schema.UnfriendRequest,
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
