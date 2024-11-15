package schema

type SetAvatarRequest struct {
	DataBase64 string `json:"dataBase64"`
}

type SetDisplayNameRequest struct {
	DisplayName string `json:"displayName"`
}
