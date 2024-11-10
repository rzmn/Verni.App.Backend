package users

import (
	"net/http"
	"verni/internal/common"
	usersController "verni/internal/controllers/users"
	httpserver "verni/internal/http-server"
	"verni/internal/http-server/responses"
	"verni/internal/services/logging"
)

type defaultRequestsHandler struct {
	controller usersController.Controller
	logger     logging.Service
}

func (c *defaultRequestsHandler) GetUsers(
	subject httpserver.UserId,
	request GetUsersRequest,
	success func(httpserver.StatusCode, responses.Response[[]httpserver.User]),
	failure func(httpserver.StatusCode, responses.Response[responses.Error]),
) {
	users, err := c.controller.Get(common.Map(request.Ids, func(id httpserver.UserId) usersController.UserId {
		return usersController.UserId(id)
	}), usersController.UserId(subject))
	if err != nil {
		switch err.Code {
		default:
			c.logger.LogError("getUsers request %v failed with unknown err: %v", request, err)
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
	success(http.StatusOK, responses.Success(common.Map(users, mapUser)))
}

func mapUser(user usersController.User) httpserver.User {
	return httpserver.User{
		Id:           httpserver.UserId(user.Id),
		DisplayName:  user.DisplayName,
		AvatarId:     (*httpserver.ImageId)(user.AvatarId),
		FriendStatus: httpserver.FriendStatus(user.FriendStatus),
	}
}
