package schema

type GetUsersRequest struct {
	Ids []UserId `json:"ids"`
}

type SearchUsersRequest struct {
	Query string `json:"query"`
}
