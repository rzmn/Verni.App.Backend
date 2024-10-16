package storage_test

import (
	"log"
	"math/rand"
	"testing"
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
	return nil
	// storage, err := ydbStorage.New(os.Getenv("YDB_TEST_ENDPOINT"), "./ydbStorage/key.json")
	// if err != nil {
	// 	t.Fatalf("%v", err)
	// }
	// _s = &storage
	// return storage
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
	newDisplayName := "newDisplayName"
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

func TestStoreAndRemoveFriendRequest(t *testing.T) {
	s := getStorage(t)
	sender := randomUid()
	target := randomUid()

	if err := s.StoreFriendRequest(sender, target); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	exists, err := s.HasFriendRequest(sender, target)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !exists {
		t.Fatalf("exists should be true")
	}
	exists, err = s.HasFriendRequest(target, sender)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if exists {
		t.Fatalf("exists should be false")
	}
	if err := s.RemoveFriendRequest(sender, target); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	exists, err = s.HasFriendRequest(sender, target)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if exists {
		t.Fatalf("exists should be false")
	}
}

func TestHasFriendRequestEmpty(t *testing.T) {
	s := getStorage(t)
	sender := randomUid()
	target := randomUid()
	exists, err := s.HasFriendRequest(sender, target)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if exists {
		t.Fatalf("exists should be false")
	}
}

func TestGetIncomingRequests(t *testing.T) {
	s := getStorage(t)
	sender := randomUid()
	target := randomUid()

	if err := s.StoreFriendRequest(sender, target); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	incoming, err := s.GetIncomingRequests(target)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(incoming) != 1 || incoming[0] != sender {
		t.Fatalf("incoming should have only sender")
	}
	incoming, err = s.GetIncomingRequests(sender)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(incoming) != 0 {
		t.Fatalf("incoming should be empty")
	}
}

func TestGetIncomingRequestsEmpty(t *testing.T) {
	s := getStorage(t)
	uid := randomUid()
	incoming, err := s.GetIncomingRequests(uid)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(incoming) != 0 {
		t.Fatalf("incoming should be empty")
	}
}

func TestGetPendingRequests(t *testing.T) {
	s := getStorage(t)
	sender := randomUid()
	target := randomUid()

	if err := s.StoreFriendRequest(sender, target); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	pending, err := s.GetPendingRequests(sender)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(pending) != 1 || pending[0] != target {
		t.Fatalf("pending should have only target")
	}
	pending, err = s.GetPendingRequests(target)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(pending) != 0 {
		t.Fatalf("pending should be empty")
	}
}

func TestGetPendingRequestsEmpty(t *testing.T) {
	s := getStorage(t)
	uid := randomUid()
	incoming, err := s.GetPendingRequests(uid)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(incoming) != 0 {
		t.Fatalf("incoming should be empty")
	}
}

func TestStoreAndRemoveFriendship(t *testing.T) {
	s := getStorage(t)
	friendA := randomUid()
	friendB := randomUid()

	if err := s.StoreFriendship(friendA, friendB); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	has, err := s.HasFriendship(friendA, friendB)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !has {
		t.Fatalf("has should be true")
	}
	has, err = s.HasFriendship(friendB, friendA)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !has {
		t.Fatalf("has should be true")
	}
	if err := s.RemoveFriendship(friendA, friendB); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	has, err = s.HasFriendship(friendA, friendB)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if has {
		t.Fatalf("has should be false")
	}
	has, err = s.HasFriendship(friendB, friendA)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if has {
		t.Fatalf("has should be false")
	}
}

func TestHasFriendshipEmpty(t *testing.T) {
	s := getStorage(t)
	friendA := randomUid()
	friendB := randomUid()
	has, err := s.HasFriendship(friendA, friendB)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if has {
		t.Fatalf("has should be false")
	}
	has, err = s.HasFriendship(friendB, friendA)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if has {
		t.Fatalf("has should be false")
	}
}

func TestGetFriends(t *testing.T) {
	s := getStorage(t)
	friendA := randomUid()
	friendB := randomUid()

	if err := s.StoreFriendship(friendA, friendB); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	friendsOfA, err := s.GetFriends(friendA)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(friendsOfA) != 1 || friendsOfA[0] != friendB {
		t.Fatalf("friendsOfA should contain only friendB")
	}
	friendsOfB, err := s.GetFriends(friendB)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(friendsOfB) != 1 || friendsOfB[0] != friendA {
		t.Fatalf("friendsOfB should contain only friendA")
	}
	if err := s.RemoveFriendship(friendB, friendA); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	friendsOfA, err = s.GetFriends(friendA)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(friendsOfA) != 0 {
		t.Fatalf("friendsOfA should be empty")
	}
	friendsOfB, err = s.GetFriends(friendB)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(friendsOfB) != 0 {
		t.Fatalf("friendsOfB should be empty")
	}
}

func TestGetFriendsEmpty(t *testing.T) {
	s := getStorage(t)
	uid := randomUid()
	friends, err := s.GetFriends(uid)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(friends) != 0 {
		t.Fatalf("friends should be empty")
	}
}

func TestGetUsers(t *testing.T) {
	s := getStorage(t)
	me := randomUid()
	meEmail := randomEmail()
	mySubscriber := randomUid()
	mySubscriberEmail := randomEmail()
	mySubscription := randomUid()
	mySubscriptionEmail := randomEmail()
	myFriend := randomUid()
	myFriendEmail := randomEmail()
	randomUser := randomUid()
	randomUserEmail := randomEmail()
	userWhoDoesNotExist := randomUid()

	usersToCreate := []storage.UserId{me, mySubscriber, mySubscription, myFriend, randomUser}
	emailsToCreate := []string{meEmail, mySubscriberEmail, mySubscriptionEmail, myFriendEmail, randomUserEmail}
	for i := 0; i < len(usersToCreate); i++ {
		pwd := randomPassword()
		if err := s.StoreCredentials(usersToCreate[i], storage.UserCredentials{Email: emailsToCreate[i], Password: pwd}); err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
	}
	if err := s.StoreFriendRequest(me, mySubscription); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if err := s.StoreFriendRequest(mySubscriber, me); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if err := s.StoreFriendship(me, myFriend); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	users, err := s.GetUsers(me, []storage.UserId{me, mySubscriber, mySubscription, myFriend, randomUser, userWhoDoesNotExist})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(usersToCreate) != len(users) {
		t.Fatalf("should get all created users, found %v", users)
	}
	passedUsers := map[storage.UserId]bool{}
	for i := 0; i < len(users); i++ {
		if users[i].Id == me && users[i].FriendStatus == storage.FriendStatusMe {
			passedUsers[me] = true
		}
		if users[i].Id == mySubscriber && users[i].FriendStatus == storage.FriendStatusIncomingRequest {
			passedUsers[mySubscriber] = true
		}
		if users[i].Id == mySubscription && users[i].FriendStatus == storage.FriendStatusOutgoingRequest {
			passedUsers[mySubscription] = true
		}
		if users[i].Id == myFriend && users[i].FriendStatus == storage.FriendStatusFriends {
			passedUsers[myFriend] = true
		}
		if users[i].Id == randomUser && users[i].FriendStatus == storage.FriendStatusNo {
			passedUsers[randomUser] = true
		}
	}
	if len(passedUsers) != len(users) {
		t.Fatalf("should pass all created users, found %v", passedUsers)
	}
}

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

func TestDealsAndCounterparties(t *testing.T) {
	s := getStorage(t)
	counterparty1 := randomUid()
	counterparty2 := randomUid()
	cost1 := int64(456)
	cost2 := int64(888)
	currency := uuid.New().String()

	deal1 := storage.Deal{
		Timestamp: 123,
		Details:   uuid.New().String(),
		Cost:      cost1,
		Currency:  currency,
		Spendings: []storage.Spending{
			{
				UserId: counterparty1,
				Cost:   cost1,
			},
			{
				UserId: counterparty2,
				Cost:   -cost1,
			},
		},
	}
	_, err := s.InsertDeal(deal1)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	deal2 := storage.Deal{
		Timestamp: 123,
		Details:   uuid.New().String(),
		Cost:      cost2,
		Currency:  currency,
		Spendings: []storage.Spending{
			{
				UserId: counterparty2,
				Cost:   -cost2 / 2,
			},
			{
				UserId: counterparty1,
				Cost:   cost2 / 2,
			},
		},
	}
	_, err = s.InsertDeal(deal2)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	deals, err := s.GetDeals(counterparty1, counterparty2)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(deals) != 2 {
		t.Fatalf("should be 2 deals, found: %v", deals)
	} else {
		log.Printf("deals ok: %v\n", deals)
	}
	counterparties, err := s.GetCounterparties(counterparty1)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(counterparties) != 1 || counterparties[0].Counterparty != string(counterparty2) || counterparties[0].Balance[currency] != (cost1+cost2/2) {
		t.Fatalf("unexpected counterparty, found: %v", counterparties)
	}
	counterparties, err = s.GetCounterparties(counterparty2)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(counterparties) != 1 || counterparties[0].Counterparty != string(counterparty1) || counterparties[0].Balance[currency] != -(cost1+cost2/2) {
		t.Fatalf("unexpected counterparty, found: %v", counterparties)
	}
}

func TestInsertAndRemoveDeal(t *testing.T) {
	s := getStorage(t)
	counterparty1 := randomUid()
	counterparty2 := randomUid()
	cost := int64(456)
	currency := uuid.New().String()

	deal := storage.Deal{
		Timestamp: 123,
		Details:   uuid.New().String(),
		Cost:      cost,
		Currency:  currency,
		Spendings: []storage.Spending{
			{
				UserId: counterparty1,
				Cost:   cost,
			},
			{
				UserId: counterparty2,
				Cost:   -cost,
			},
		},
	}
	_, err := s.InsertDeal(deal)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	deals, err := s.GetDeals(counterparty1, counterparty2)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(deals) != 1 {
		t.Fatalf("deals len should be 1, found: %v", deals)
	}
	dealFromDb, err := s.GetDeal(deals[0].Id)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if dealFromDb == nil {
		t.Fatalf("deal should exists: %v", err)
	}
	if err := s.RemoveDeal(deals[0].Id); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	dealFromDb, err = s.GetDeal(deals[0].Id)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if dealFromDb != nil {
		t.Fatalf("deal should not exists: %v", err)
	}
}

func TestGetCounterpartiesForDeal(t *testing.T) {
	s := getStorage(t)
	counterparty1 := randomUid()
	counterparty2 := randomUid()
	cost := int64(456)
	currency := uuid.New().String()

	deal := storage.Deal{
		Timestamp: 123,
		Details:   uuid.New().String(),
		Cost:      cost,
		Currency:  currency,
		Spendings: []storage.Spending{
			{
				UserId: counterparty1,
				Cost:   cost,
			},
			{
				UserId: counterparty2,
				Cost:   -cost,
			},
		},
	}
	_, err := s.InsertDeal(deal)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	deals, err := s.GetDeals(counterparty1, counterparty2)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(deals) != 1 {
		t.Fatalf("deals len should be 1, found: %v", deals)
	}
	counterparties, err := s.GetCounterpartiesForDeal(deals[0].Id)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(counterparties) != 2 {
		t.Fatalf("counterparties len should be 2, found: %v", counterparties)
	}
	passedUsers := map[storage.UserId]bool{}
	for i := 0; i < len(counterparties); i++ {
		if counterparties[i] == counterparty1 {
			passedUsers[counterparties[i]] = true
		}
		if counterparties[i] == counterparty2 {
			passedUsers[counterparties[i]] = true
		}
	}
	if len(passedUsers) != 2 {
		t.Fatalf("passedUsers len should be 2, found: %v", passedUsers)
	}
}
