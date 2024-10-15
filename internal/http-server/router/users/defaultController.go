package users

import (
	"accounty/internal/common"
	"accounty/internal/storage"
	"log"
)

type defaultController struct {
	storage storage.Storage
}

func (c *defaultController) Get(ids []UserId, sender UserId) ([]User, *common.CodeBasedError[GetUsersErrorCode]) {
	const op = "users.defaultController.Get"
	log.Printf("%s: start[sender=%s, ids=%v]", op, sender, ids)
	users, err := c.storage.GetUsers(storage.UserId(sender), common.Map(ids, func(id UserId) storage.UserId {
		return storage.UserId(id)
	}))
	if err != nil {
		log.Printf("%s: cannot read from db err: %v", op, err)
		return []User{}, common.NewErrorWithDescription(GetUsersErrorInternal, err.Error())
	}
	log.Printf("%s: success[sender=%s, ids=%v]", op, sender, ids)
	return common.Map(users, func(user storage.User) User {
		return User(user)
	}), nil
}

func (c *defaultController) Search(query string, sender UserId) ([]User, *common.CodeBasedError[SearchUsersErrorCode]) {
	const op = "users.defaultController.Search"
	log.Printf("%s: start[sender=%s, query=%v]", op, sender, query)
	users, err := c.storage.SearchUsers(storage.UserId(sender), query)
	if err != nil {
		log.Printf("%s: cannot read from db err: %v", op, err)
		return []User{}, common.NewErrorWithDescription(SearchUsersErrorInternal, err.Error())
	}
	log.Printf("%s: success[sender=%s, query=%v]", op, sender, query)
	return common.Map(users, func(user storage.User) User {
		return User(user)
	}), nil
}
