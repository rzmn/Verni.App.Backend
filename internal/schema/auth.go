package schema

type SignupRequest struct {
	Credentials Credentials `json:"credentials"`
}

type LoginRequest struct {
	Credentials Credentials `json:"credentials"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type UpdateEmailRequest struct {
	Email string `json:"email"`
}

type UpdatePasswordRequest struct {
	OldPassword string `json:"old"`
	NewPassword string `json:"new"`
}

type RegisterForPushNotificationsRequest struct {
	Token string `json:"token"`
}
