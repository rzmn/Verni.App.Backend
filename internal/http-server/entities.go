package httpserver

type UserId string
type ExpenseId string
type ImageId string
type FriendStatus int
type Cost int64
type Currency string

const (
	FriendStatusNo = iota
	FriendStatusSubscriber
	FriendStatusSubscription
	FriendStatusFriend
	FriendStatusMe
)

type Image struct {
	Id         ImageId `json:"id"`
	Base64Data *string `json:"base64"`
}

type User struct {
	Id           UserId       `json:"id"`
	DisplayName  string       `json:"displayName"`
	AvatarId     *ImageId     `json:"avatarId"`
	FriendStatus FriendStatus `json:"friendStatus"`
}

type Profile struct {
	User          User   `json:"user"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"emailVerified"`
}

type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ExpenseAttachment struct {
	ImageId *ImageId
}

type Session struct {
	Id           UserId `json:"id"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type Expense struct {
	Timestamp   int64               `json:"timestamp"`
	Details     string              `json:"details"`
	Total       Cost                `json:"total"`
	Attachments []ExpenseAttachment `json:"attachments"`
	Currency    Currency            `json:"currency"`
	Shares      []ShareOfExpense    `json:"shares"`
}

type IdentifiableExpense struct {
	Expense
	Id ExpenseId `json:"id"`
}

type ShareOfExpense struct {
	UserId UserId `json:"userId"`
	Cost   Cost   `json:"cost"`
}

type Balance struct {
	Counterparty string            `json:"counterparty"`
	Currencies   map[Currency]Cost `json:"currencies"`
}
