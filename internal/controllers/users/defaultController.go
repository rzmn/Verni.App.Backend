package users

import (
	"log"
	"verni/internal/common"
	"verni/internal/repositories/users"
)

type defaultController struct {
	repository Repository
}

func (c *defaultController) Get(ids []UserId, sender UserId) ([]User, *common.CodeBasedError[GetUsersErrorCode]) {
	const op = "users.defaultController.Get"
	log.Printf("%s: start[sender=%s, ids=%v]", op, sender, ids)
	usersFromRepository, err := c.repository.GetUsers(common.Map(ids, func(id UserId) users.UserId {
		return users.UserId(id)
	}))
	if err != nil {
		log.Printf("%s: cannot read from db err: %v", op, err)
		return []User{}, common.NewErrorWithDescription(GetUsersErrorInternal, err.Error())
	}
	log.Printf("%s: success[sender=%s, ids=%v]", op, sender, ids)
	return common.Map(usersFromRepository, mapUser), nil
}

func (c *defaultController) Search(query string, sender UserId) ([]User, *common.CodeBasedError[SearchUsersErrorCode]) {
	const op = "users.defaultController.Search"
	log.Printf("%s: start[sender=%s, query=%v]", op, sender, query)
	users, err := c.repository.SearchUsers(query)
	if err != nil {
		log.Printf("%s: cannot read from db err: %v", op, err)
		return []User{}, common.NewErrorWithDescription(SearchUsersErrorInternal, err.Error())
	}
	log.Printf("%s: success[sender=%s, query=%v]", op, sender, query)
	return common.Map(users, mapUser), nil
}

func mapUser(users.User) User {
	return User{}
}
