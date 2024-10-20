package storage_test

import (
	"encoding/json"
	"io"
	"log"
	"math/rand"
	"os"
	"testing"
	"verni/internal/common"
	"verni/internal/storage"

	"github.com/google/uuid"
)

var (
	_s *storage.Storage
)

func randomUid() storage.UserId {
	return storage.UserId(uuid.New().String())
}

func randomPassword() string {
	return uuid.New().String()
}

func randomEmail() string {
	const characters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	email := make([]byte, 15)
	for i := range email {
		email[i] = characters[rand.Intn(len(characters))]
	}
	return string(email) + "@x.com"
}

func getStorage(t *testing.T) storage.Storage {
	if _s != nil {
		return *_s
	}

	common.RegisterRelativePathRoot(os.Getenv("VERNI_PROJECT_ROOT"))
	configFile, err := os.Open(common.AbsolutePath("./config/test/ydb_test_environment.json"))
	if err != nil {
		t.Fatalf("failed to open config file: %s", err)
	}
	defer configFile.Close()
	configData, err := io.ReadAll(configFile)
	if err != nil {
		t.Fatalf("failed to read config file: %s", err)
	}
	var config storage.YDBConfig
	json.Unmarshal([]byte(configData), &config)
	log.Printf("initializing with config %v", config)
	db, err := storage.YDB(config)
	if err != nil {
		t.Fatalf("failed to init db err: %v", err)
	}
	_s = &db
	return db
}

func TestIsUserExistsFalse(t *testing.T) {
	s := getStorage(t)
	uid := randomUid()
	exists, err := s.IsUserExists(uid)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if exists {
		t.Fatalf("unexpected exists=true")
	}
}

func TestIsUserExistsTrue(t *testing.T) {
	s := getStorage(t)
	uid := randomUid()
	pwd := randomPassword()
	email := randomEmail()
	credentials := storage.UserCredentials{Email: email, Password: pwd}
	if err := s.StoreCredentials(uid, credentials); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	exists, err := s.IsUserExists(uid)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !exists {
		t.Fatalf("unexpected exists=false")
	}
}

func TestStorePushToken(t *testing.T) {
	s := getStorage(t)
	uid := randomUid()
	token := uuid.New().String()
	if err := s.StorePushToken(uid, token); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	tokenFromDb, err := s.GetPushToken(uid)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if tokenFromDb == nil {
		t.Fatalf("unexpected nil")
	}
	if *tokenFromDb != token {
		t.Fatalf("should be equal")
	}
}

func TestGetAccountInfo(t *testing.T) {
	s := getStorage(t)
	uid := randomUid()
	pwd := randomPassword()
	email := randomEmail()
	credentials := storage.UserCredentials{Email: email, Password: pwd}
	if err := s.StoreCredentials(uid, credentials); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	info, err := s.GetAccountInfo(uid)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if info == nil {
		t.Fatalf("unexpected exists=false")
	}
	log.Printf("info: %v\n", *info)
}

func TestGetUserId(t *testing.T) {
	s := getStorage(t)
	uid := randomUid()
	pwd := randomPassword()
	email := randomEmail()
	credentials := storage.UserCredentials{Email: email, Password: pwd}
	if err := s.StoreCredentials(uid, credentials); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	uidFromQuery, err := s.GetUserId(email)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if uidFromQuery == nil {
		t.Fatalf("unexpected nil")
	}
	if *uidFromQuery != uid {
		t.Fatalf("unexpected id mismatch, found %v", uidFromQuery)
	}
}

func TestGetUserIdEmpty(t *testing.T) {
	s := getStorage(t)
	email := randomEmail()
	uidFromQuery, err := s.GetUserId(email)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if uidFromQuery != nil {
		t.Fatalf("unexpected non-nil found %v", uidFromQuery)
	}
}

func TestCheckCredentialsTrue(t *testing.T) {
	s := getStorage(t)
	uid := randomUid()
	pwd := randomPassword()
	email := randomEmail()
	credentials := storage.UserCredentials{Email: email, Password: pwd}
	if err := s.StoreCredentials(uid, credentials); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	passed, err := s.CheckCredentials(credentials)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !passed {
		t.Fatalf("unexpected passed=false")
	}
}

func TestUpdateDisplayName(t *testing.T) {
	s := getStorage(t)
	uid := randomUid()
	pwd := randomPassword()
	email := randomEmail()
	credentials := storage.UserCredentials{Email: email, Password: pwd}
	if err := s.StoreCredentials(uid, credentials); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	info, err := s.GetAccountInfo(uid)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if info == nil {
		t.Fatalf("no account info")
	}
	if info.User.DisplayName != email {
		t.Fatalf("initial display name should be a email, found %s", info.User.DisplayName)
	}
	newDisplayName := "newDisplayName 💅🇷🇺"
	if err := s.StoreDisplayName(uid, newDisplayName); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	info, err = s.GetAccountInfo(uid)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if info == nil {
		t.Fatalf("no account info")
	}
	if info.User.DisplayName != newDisplayName {
		t.Fatalf("initial display name should be %s, found %s", newDisplayName, info.User.DisplayName)
	}
}

func TestUpdateAvatar(t *testing.T) {
	s := getStorage(t)
	uid := randomUid()
	pwd := randomPassword()
	email := randomEmail()
	credentials := storage.UserCredentials{Email: email, Password: pwd}
	if err := s.StoreCredentials(uid, credentials); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	info, err := s.GetAccountInfo(uid)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if info == nil {
		t.Fatalf("no account info")
	}
	if info.User.Avatar.Id != nil {
		t.Fatalf("unexpected non-nil avatar, found %s", *info.User.Avatar.Id)
	}
	newAvatar := "xxx"
	_, err = s.StoreAvatarBase64(uid, newAvatar)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	info, err = s.GetAccountInfo(uid)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if info == nil {
		t.Fatalf("no account info")
	}
	if info.User.Avatar.Id == nil {
		t.Fatalf("new avatar id should not be nil")
	}
	avatars, err := s.GetAvatarsBase64([]storage.AvatarId{*info.User.Avatar.Id})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(avatars) != 1 {
		t.Fatalf("avatars len should be 1, found: %v", avatars)
	}
	if *avatars[*info.User.Avatar.Id].Base64Data != newAvatar {
		t.Fatalf("avatars data did not match, found: %v-%v", *avatars[*info.User.Avatar.Id].Base64Data, newAvatar)
	}
}

func TestCheckCredentialsFalse(t *testing.T) {
	s := getStorage(t)
	uid := randomUid()
	pwd := randomPassword()
	email := randomEmail()
	credentials := storage.UserCredentials{Email: email, Password: pwd}
	if err := s.StoreCredentials(uid, credentials); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	credentials.Password = uuid.New().String()
	passed, err := s.CheckCredentials(credentials)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if passed {
		t.Fatalf("unexpected passed=true")
	}
}

func TestCheckCredentialsEmpty(t *testing.T) {
	s := getStorage(t)
	pwd := randomPassword()
	email := randomEmail()
	credentials := storage.UserCredentials{Email: email, Password: pwd}
	passed, err := s.CheckCredentials(credentials)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if passed {
		t.Fatalf("unexpected passed=true")
	}
}

func TestStoreAndRemoveRefreshToken(t *testing.T) {
	s := getStorage(t)
	uid := randomUid()
	token := uuid.New().String()
	if err := s.StoreRefreshToken(token, uid); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	tokenFromStorage, err := s.GetRefreshToken(uid)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if token != *tokenFromStorage {
		t.Fatalf("tokens are not equal")
	}
	if err := s.RemoveRefreshToken(uid); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	nilTokenFromStorage, err := s.GetRefreshToken(uid)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if nilTokenFromStorage != nil {
		t.Fatalf("token should be nil")
	}
}

func TestGetRefreshTokenEmpty(t *testing.T) {
	s := getStorage(t)
	uid := randomUid()
	token, err := s.GetRefreshToken(uid)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if token != nil {
		t.Fatalf("token should be nil")
	}
}

// func TestGetUsers(t *testing.T) {
// 	s := getStorage(t)
// 	me := randomUid()
// 	meEmail := randomEmail()
// 	mySubscriber := randomUid()
// 	mySubscriberEmail := randomEmail()
// 	mySubscription := randomUid()
// 	mySubscriptionEmail := randomEmail()
// 	myFriend := randomUid()
// 	myFriendEmail := randomEmail()
// 	randomUser := randomUid()
// 	randomUserEmail := randomEmail()
// 	userWhoDoesNotExist := randomUid()

// 	usersToCreate := []storage.UserId{me, mySubscriber, mySubscription, myFriend, randomUser}
// 	emailsToCreate := []string{meEmail, mySubscriberEmail, mySubscriptionEmail, myFriendEmail, randomUserEmail}
// 	for i := 0; i < len(usersToCreate); i++ {
// 		pwd := randomPassword()
// 		if err := s.StoreCredentials(usersToCreate[i], storage.UserCredentials{Email: emailsToCreate[i], Password: pwd}); err != nil {
// 			t.Fatalf("unexpected err: %v", err)
// 		}
// 	}
// 	if err := s.StoreFriendRequest(me, mySubscription); err != nil {
// 		t.Fatalf("unexpected err: %v", err)
// 	}
// 	if err := s.StoreFriendRequest(mySubscriber, me); err != nil {
// 		t.Fatalf("unexpected err: %v", err)
// 	}
// 	if err := s.StoreFriendship(me, myFriend); err != nil {
// 		t.Fatalf("unexpected err: %v", err)
// 	}
// 	users, err := s.GetUsers(me, []storage.UserId{me, mySubscriber, mySubscription, myFriend, randomUser, userWhoDoesNotExist})
// 	if err != nil {
// 		t.Fatalf("unexpected err: %v", err)
// 	}
// 	if len(usersToCreate) != len(users) {
// 		t.Fatalf("should get all created users, found %v", users)
// 	}
// 	passedUsers := map[storage.UserId]bool{}
// 	for i := 0; i < len(users); i++ {
// 		if users[i].Id == me && users[i].FriendStatus == storage.FriendStatusMe {
// 			passedUsers[me] = true
// 		}
// 		if users[i].Id == mySubscriber && users[i].FriendStatus == storage.FriendStatusIncomingRequest {
// 			passedUsers[mySubscriber] = true
// 		}
// 		if users[i].Id == mySubscription && users[i].FriendStatus == storage.FriendStatusOutgoingRequest {
// 			passedUsers[mySubscription] = true
// 		}
// 		if users[i].Id == myFriend && users[i].FriendStatus == storage.FriendStatusFriends {
// 			passedUsers[myFriend] = true
// 		}
// 		if users[i].Id == randomUser && users[i].FriendStatus == storage.FriendStatusNo {
// 			passedUsers[randomUser] = true
// 		}
// 	}
// 	if len(passedUsers) != len(users) {
// 		t.Fatalf("should pass all created users, found %v", passedUsers)
// 	}
// }

func TestGetUsersEmpty(t *testing.T) {
	s := getStorage(t)
	uid := randomUid()
	users, err := s.GetUsers(uid, []storage.UserId{uid})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(users) != 0 {
		t.Fatalf("users should be empty, found %v", users)
	}
}

func TestSearchUsers(t *testing.T) {
	s := getStorage(t)
	me := randomUid()
	meEmail := randomEmail()
	otherUser := randomUid()
	otherUserEmail := randomEmail()

	usersToCreate := []storage.UserId{me, otherUser}
	emailsToCreate := []string{meEmail, otherUserEmail}
	for i := 0; i < len(usersToCreate); i++ {
		pwd := uuid.New().String()
		if err := s.StoreCredentials(usersToCreate[i], storage.UserCredentials{Email: emailsToCreate[i], Password: pwd}); err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
	}
	result1, err := s.SearchUsers(me, string(otherUserEmail))
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(result1) != 1 || result1[0].Id != otherUser {
		t.Fatalf("result1 should contain only otherUser, found: %v", result1)
	}
	result2, err := s.SearchUsers(me, string(otherUserEmail)[0:len(otherUserEmail)-3])
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(result2) != 1 || result2[0].Id != otherUser {
		t.Fatalf("result2 should contain only otherUser for req %s, found: %v", string(otherUserEmail)[0:len(otherUserEmail)-3], result2)
	}
}

func TestSearchUsersEmpty(t *testing.T) {
	s := getStorage(t)
	me := randomUid()
	result, err := s.SearchUsers(me, "")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("result should be empty, found: %v", result)
	}
	result, err = s.SearchUsers(me, "+++")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("result should be empty, found: %v", result)
	}
}
