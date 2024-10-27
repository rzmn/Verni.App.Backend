package friends_test

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"testing"
	"verni/internal/common"
	"verni/internal/db"
	"verni/internal/repositories/friends"

	"github.com/google/uuid"
)

var (
	database db.DB
)

func TestMain(m *testing.M) {
	database = func() db.DB {
		configFile, err := os.Open(common.AbsolutePath("./config/test/postgres_storage.json"))
		if err != nil {
			log.Fatalf("failed to open config file: %s", err)
		}
		defer configFile.Close()
		configData, err := io.ReadAll(configFile)
		if err != nil {
			log.Fatalf("failed to read config file: %s", err)
		}
		var config db.PostgresConfig
		json.Unmarshal([]byte(configData), &config)
		db, err := db.Postgres(config)
		if err != nil {
			log.Fatalf("failed to init db err: %v", err)
		}
		return db
	}()
	code := m.Run()

	os.Exit(code)
}

func init() {
	root, present := os.LookupEnv("VERNI_PROJECT_ROOT")
	if present {
		common.RegisterRelativePathRoot(root)
	} else {
		log.Fatalf("project root not found")
	}
}

func randomUid() friends.UserId {
	return friends.UserId(uuid.New().String())
}

func TestSubscribtion(t *testing.T) {
	repository := friends.PostgresRepository(database)
	subscriber := randomUid()
	subscription := randomUid()

	subsctiptionTransaction := repository.StoreFriendRequest(subscriber, subscription)
	if err := subsctiptionTransaction.Perform(); err != nil {
		t.Fatalf("failed to perform `subsctiptionTransaction` err: %v", err)
	}
	hasRequestFromSubscriberToSubscription, err := repository.HasFriendRequest(subscriber, subscription)
	if err != nil {
		t.Fatalf("failed to get `hasRequestFromSubscriberToSubscription` err: %v", err)
	}
	if !hasRequestFromSubscriberToSubscription {
		t.Fatalf("`hasRequestFromSubscriberToSubscription` should be true")
	}
	hasRequestFromSubscriptionToSubscriber, err := repository.HasFriendRequest(subscription, subscriber)
	if err != nil {
		t.Fatalf("failed to get `hasRequestFromSubscriptionToSubscriber` err: %v", err)
	}
	if hasRequestFromSubscriptionToSubscriber {
		t.Fatalf("`hasRequestFromSubscriptionToSubscriber` should be false")
	}
	subscriptionsOfSubscriber, err := repository.GetSubscriptions(subscriber)
	if err != nil {
		t.Fatalf("failed to get `subscriptionsOfSubscriber` err: %v", err)
	}
	if len(subscriptionsOfSubscriber) != 1 || subscriptionsOfSubscriber[0] != subscription {
		t.Fatalf("`subscriptionsOfSubscriber` should contain %s only, found %v", subscription, subscriptionsOfSubscriber)
	}
	subscriptionsOfSubscription, err := repository.GetSubscriptions(subscription)
	if err != nil {
		t.Fatalf("failed to get `subscriptionsOfSubscription` err: %v", err)
	}
	if len(subscriptionsOfSubscription) != 0 {
		t.Fatalf("`subscriptionsOfSubscription` be empty, found %v", subscriptionsOfSubscription)
	}
	subscribersOfSubscriber, err := repository.GetSubscribers(subscriber)
	if err != nil {
		t.Fatalf("failed to get `subscribersOfSubscriber` err: %v", err)
	}
	if len(subscribersOfSubscriber) != 0 {
		t.Fatalf("`subscribersOfSubscriber` should be empty, found %v", subscribersOfSubscriber)
	}
	subscribersOfSubscription, err := repository.GetSubscribers(subscription)
	if err != nil {
		t.Fatalf("failed to get `subscribersOfSubscription` err: %v", err)
	}
	if len(subscribersOfSubscription) != 1 || subscribersOfSubscription[0] != subscriber {
		t.Fatalf("`subscribersOfSubscription` should contain %s only, found %v", subscriber, subscribersOfSubscription)
	}
	friendsOfSubscriber, err := repository.GetFriends(subscriber)
	if err != nil {
		t.Fatalf("failed to get `friendsOfSubscriber` err: %v", err)
	}
	if len(friendsOfSubscriber) != 0 {
		t.Fatalf("`friendsOfSubscriber` should be empty, found %v", friendsOfSubscriber)
	}
	friendsOfSubscription, err := repository.GetFriends(subscription)
	if err != nil {
		t.Fatalf("failed to get `friendsOfSubscription` err: %v", err)
	}
	if len(friendsOfSubscription) != 0 {
		t.Fatalf("`friendsOfSubscription` should be empty, found %v", friendsOfSubscription)
	}
	if err := subsctiptionTransaction.Rollback(); err != nil {
		t.Fatalf("failed to rollback `subsctiptionTransaction` err: %v", err)
	}
	hasRequestFromSubscriberToSubscription, err = repository.HasFriendRequest(subscriber, subscription)
	if err != nil {
		t.Fatalf("[after rollback] failed to get `hasRequestFromSubscriberToSubscription` err: %v", err)
	}
	if hasRequestFromSubscriberToSubscription {
		t.Fatalf("[after rollback] `hasRequestFromSubscriberToSubscription` should be false")
	}
	hasRequestFromSubscriptionToSubscriber, err = repository.HasFriendRequest(subscription, subscriber)
	if err != nil {
		t.Fatalf("[after rollback] failed to get `hasRequestFromSubscriptionToSubscriber` err: %v", err)
	}
	if hasRequestFromSubscriptionToSubscriber {
		t.Fatalf("[after rollback] `hasRequestFromSubscriptionToSubscriber` should be false")
	}
	subscriptionsOfSubscriber, err = repository.GetSubscriptions(subscriber)
	if err != nil {
		t.Fatalf("[after rollback] failed to get `subscriptionsOfSubscriber` err: %v", err)
	}
	if len(subscriptionsOfSubscriber) != 0 {
		t.Fatalf("[after rollback] `subscriptionsOfSubscriber` should be empty, found %v", subscriptionsOfSubscriber)
	}
	subscriptionsOfSubscription, err = repository.GetSubscriptions(subscription)
	if err != nil {
		t.Fatalf("[after rollback] failed to get `subscriptionsOfSubscription` err: %v", err)
	}
	if len(subscriptionsOfSubscription) != 0 {
		t.Fatalf("[after rollback] `subscriptionsOfSubscription` be empty, found %v", subscriptionsOfSubscription)
	}
	subscribersOfSubscriber, err = repository.GetSubscribers(subscriber)
	if err != nil {
		t.Fatalf("[after rollback] failed to get `subscribersOfSubscriber` err: %v", err)
	}
	if len(subscribersOfSubscriber) != 0 {
		t.Fatalf("[after rollback] `subscribersOfSubscriber` should be empty, found %v", subscribersOfSubscriber)
	}
	subscribersOfSubscription, err = repository.GetSubscribers(subscription)
	if err != nil {
		t.Fatalf("[after rollback] failed to get `subscribersOfSubscription` err: %v", err)
	}
	if len(subscribersOfSubscription) != 0 {
		t.Fatalf("[after rollback] `subscribersOfSubscription` should be empty, found %v", subscribersOfSubscription)
	}
	friendsOfSubscriber, err = repository.GetFriends(subscriber)
	if err != nil {
		t.Fatalf("[after rollback] failed to get `friendsOfSubscriber` err: %v", err)
	}
	if len(friendsOfSubscriber) != 0 {
		t.Fatalf("[after rollback] `friendsOfSubscriber` should be empty, found %v", friendsOfSubscriber)
	}
	friendsOfSubscription, err = repository.GetFriends(subscription)
	if err != nil {
		t.Fatalf("[after rollback] failed to get `friendsOfSubscription` err: %v", err)
	}
	if len(friendsOfSubscription) != 0 {
		t.Fatalf("[after rollback] `friendsOfSubscription` should be empty, found %v", friendsOfSubscription)
	}
}

func TestStoreAndRemoveFriendRequest(t *testing.T) {
	repository := friends.PostgresRepository(database)
	subscriber := randomUid()
	subscription := randomUid()

	subsctiptionTransaction := repository.StoreFriendRequest(subscriber, subscription)
	if err := subsctiptionTransaction.Perform(); err != nil {
		t.Fatalf("failed to perform `subsctiptionTransaction` err: %v", err)
	}
	hasRequestFromSubscriberToSubscription, err := repository.HasFriendRequest(subscriber, subscription)
	if err != nil {
		t.Fatalf("failed to get `hasRequestFromSubscriberToSubscription` err: %v", err)
	}
	if !hasRequestFromSubscriberToSubscription {
		t.Fatalf("`hasRequestFromSubscriberToSubscription` should be true")
	}
	hasRequestFromSubscriptionToSubscriber, err := repository.HasFriendRequest(subscription, subscriber)
	if err != nil {
		t.Fatalf("failed to get `hasRequestFromSubscriptionToSubscriber` err: %v", err)
	}
	if hasRequestFromSubscriptionToSubscriber {
		t.Fatalf("`hasRequestFromSubscriptionToSubscriber` should be false")
	}
	subscriptionsOfSubscriber, err := repository.GetSubscriptions(subscriber)
	if err != nil {
		t.Fatalf("failed to get `subscriptionsOfSubscriber` err: %v", err)
	}
	if len(subscriptionsOfSubscriber) != 1 || subscriptionsOfSubscriber[0] != subscription {
		t.Fatalf("`subscriptionsOfSubscriber` should contain %s only, found %v", subscription, subscriptionsOfSubscriber)
	}
	subscriptionsOfSubscription, err := repository.GetSubscriptions(subscription)
	if err != nil {
		t.Fatalf("failed to get `subscriptionsOfSubscription` err: %v", err)
	}
	if len(subscriptionsOfSubscription) != 0 {
		t.Fatalf("`subscriptionsOfSubscription` be empty, found %v", subscriptionsOfSubscription)
	}
	subscribersOfSubscriber, err := repository.GetSubscribers(subscriber)
	if err != nil {
		t.Fatalf("failed to get `subscribersOfSubscriber` err: %v", err)
	}
	if len(subscribersOfSubscriber) != 0 {
		t.Fatalf("`subscribersOfSubscriber` should be empty, found %v", subscribersOfSubscriber)
	}
	subscribersOfSubscription, err := repository.GetSubscribers(subscription)
	if err != nil {
		t.Fatalf("failed to get `subscribersOfSubscription` err: %v", err)
	}
	if len(subscribersOfSubscription) != 1 || subscribersOfSubscription[0] != subscriber {
		t.Fatalf("`subscribersOfSubscription` should contain %s only, found %v", subscriber, subscribersOfSubscription)
	}
	friendsOfSubscriber, err := repository.GetFriends(subscriber)
	if err != nil {
		t.Fatalf("failed to get `friendsOfSubscriber` err: %v", err)
	}
	if len(friendsOfSubscriber) != 0 {
		t.Fatalf("`friendsOfSubscriber` should be empty, found %v", friendsOfSubscriber)
	}
	friendsOfSubscription, err := repository.GetFriends(subscription)
	if err != nil {
		t.Fatalf("failed to get `friendsOfSubscription` err: %v", err)
	}
	if len(friendsOfSubscription) != 0 {
		t.Fatalf("`friendsOfSubscription` should be empty, found %v", friendsOfSubscription)
	}
	removeTransaction := repository.RemoveFriendRequest(subscriber, subscription)
	if err := removeTransaction.Perform(); err != nil {
		t.Fatalf("failed to perform `removeTransaction` err: %v", err)
	}
	hasRequestFromSubscriberToSubscription, err = repository.HasFriendRequest(subscriber, subscription)
	if err != nil {
		t.Fatalf("[after remove] failed to get `hasRequestFromSubscriberToSubscription` err: %v", err)
	}
	if hasRequestFromSubscriberToSubscription {
		t.Fatalf("[after remove] `hasRequestFromSubscriberToSubscription` should be false")
	}
	hasRequestFromSubscriptionToSubscriber, err = repository.HasFriendRequest(subscription, subscriber)
	if err != nil {
		t.Fatalf("[after remove] failed to get `hasRequestFromSubscriptionToSubscriber` err: %v", err)
	}
	if hasRequestFromSubscriptionToSubscriber {
		t.Fatalf("[after remove] `hasRequestFromSubscriptionToSubscriber` should be false")
	}
	subscriptionsOfSubscriber, err = repository.GetSubscriptions(subscriber)
	if err != nil {
		t.Fatalf("[after remove] failed to get `subscriptionsOfSubscriber` err: %v", err)
	}
	if len(subscriptionsOfSubscriber) != 0 {
		t.Fatalf("[after remove] `subscriptionsOfSubscriber` should be empty, found %v", subscriptionsOfSubscriber)
	}
	subscriptionsOfSubscription, err = repository.GetSubscriptions(subscription)
	if err != nil {
		t.Fatalf("[after remove] failed to get `subscriptionsOfSubscription` err: %v", err)
	}
	if len(subscriptionsOfSubscription) != 0 {
		t.Fatalf("[after remove] `subscriptionsOfSubscription` be empty, found %v", subscriptionsOfSubscription)
	}
	subscribersOfSubscriber, err = repository.GetSubscribers(subscriber)
	if err != nil {
		t.Fatalf("[after remove] failed to get `subscribersOfSubscriber` err: %v", err)
	}
	if len(subscribersOfSubscriber) != 0 {
		t.Fatalf("[after remove] `subscribersOfSubscriber` should be empty, found %v", subscribersOfSubscriber)
	}
	subscribersOfSubscription, err = repository.GetSubscribers(subscription)
	if err != nil {
		t.Fatalf("[after remove] failed to get `subscribersOfSubscription` err: %v", err)
	}
	if len(subscribersOfSubscription) != 0 {
		t.Fatalf("[after remove] `subscribersOfSubscription` should be empty, found %v", subscribersOfSubscription)
	}
	friendsOfSubscriber, err = repository.GetFriends(subscriber)
	if err != nil {
		t.Fatalf("[after rollback] failed to get `friendsOfSubscriber` err: %v", err)
	}
	if len(friendsOfSubscriber) != 0 {
		t.Fatalf("[after rollback] `friendsOfSubscriber` should be empty, found %v", friendsOfSubscriber)
	}
	friendsOfSubscription, err = repository.GetFriends(subscription)
	if err != nil {
		t.Fatalf("[after rollback] failed to get `friendsOfSubscription` err: %v", err)
	}
	if len(friendsOfSubscription) != 0 {
		t.Fatalf("[after rollback] `friendsOfSubscription` should be empty, found %v", friendsOfSubscription)
	}
}

// func TestFriendship(t *testing.T) {
// 	repository := friends.PostgresRepository(database)
// 	subscriber := randomUid()
// 	subscription := randomUid()

// 	friendshipTransaction := []repositories.MutationWorkItem{
// 		repository.StoreFriendRequest(subscriber, subscription),
// 		repository.StoreFriendRequest(subscription, subscriber),
// 	}
// 	for i := 0; i < len(friendshipTransaction); i++ {
// 		if err := friendshipTransaction[0].Perform(); err != nil {
// 			t.Fatalf("failed to perform `friendshipTransaction[%d]` err: %v", i, err)
// 		}
// 	}
// 	hasRequestFromSubscriberToSubscription, err := repository.HasFriendRequest(subscriber, subscription)
// 	if err != nil {
// 		t.Fatalf("failed to get `hasRequestFromSubscriberToSubscription` err: %v", err)
// 	}
// 	if !hasRequestFromSubscriberToSubscription {
// 		t.Fatalf("`hasRequestFromSubscriberToSubscription` should be true")
// 	}
// 	hasRequestFromSubscriptionToSubscriber, err := repository.HasFriendRequest(subscription, subscriber)
// 	if err != nil {
// 		t.Fatalf("failed to get `hasRequestFromSubscriptionToSubscriber` err: %v", err)
// 	}
// 	if hasRequestFromSubscriptionToSubscriber {
// 		t.Fatalf("`hasRequestFromSubscriptionToSubscriber` should be false")
// 	}
// 	subscriptionsOfSubscriber, err := repository.GetSubscriptions(subscriber)
// 	if err != nil {
// 		t.Fatalf("failed to get `subscriptionsOfSubscriber` err: %v", err)
// 	}
// 	if len(subscriptionsOfSubscriber) != 1 || subscriptionsOfSubscriber[0] != subscription {
// 		t.Fatalf("`subscriptionsOfSubscriber` should contain %s only, found %v", subscription, subscriptionsOfSubscriber)
// 	}
// 	subscriptionsOfSubscription, err := repository.GetSubscriptions(subscription)
// 	if err != nil {
// 		t.Fatalf("failed to get `subscriptionsOfSubscription` err: %v", err)
// 	}
// 	if len(subscriptionsOfSubscription) != 0 {
// 		t.Fatalf("`subscriptionsOfSubscription` be empty, found %v", subscriptionsOfSubscription)
// 	}
// 	subscribersOfSubscriber, err := repository.GetSubscribers(subscriber)
// 	if err != nil {
// 		t.Fatalf("failed to get `subscribersOfSubscriber` err: %v", err)
// 	}
// 	if len(subscribersOfSubscriber) != 0 {
// 		t.Fatalf("`subscribersOfSubscriber` should be empty, found %v", subscribersOfSubscriber)
// 	}
// 	subscribersOfSubscription, err := repository.GetSubscribers(subscription)
// 	if err != nil {
// 		t.Fatalf("failed to get `subscribersOfSubscription` err: %v", err)
// 	}
// 	if len(subscribersOfSubscription) != 1 || subscribersOfSubscription[0] != subscriber {
// 		t.Fatalf("`subscribersOfSubscription` should contain %s only, found %v", subscriber, subscribersOfSubscription)
// 	}
// 	if err := subsctiptionTransaction.Rollback(); err != nil {
// 		t.Fatalf("failed to rollback `subsctiptionTransaction` err: %v", err)
// 	}
// 	hasRequestFromSubscriberToSubscription, err = repository.HasFriendRequest(subscriber, subscription)
// 	if err != nil {
// 		t.Fatalf("[after rollback] failed to get `hasRequestFromSubscriberToSubscription` err: %v", err)
// 	}
// 	if hasRequestFromSubscriberToSubscription {
// 		t.Fatalf("[after rollback] `hasRequestFromSubscriberToSubscription` should be false")
// 	}
// 	hasRequestFromSubscriptionToSubscriber, err = repository.HasFriendRequest(subscription, subscriber)
// 	if err != nil {
// 		t.Fatalf("[after rollback] failed to get `hasRequestFromSubscriptionToSubscriber` err: %v", err)
// 	}
// 	if hasRequestFromSubscriptionToSubscriber {
// 		t.Fatalf("[after rollback] `hasRequestFromSubscriptionToSubscriber` should be false")
// 	}
// 	subscriptionsOfSubscriber, err = repository.GetSubscriptions(subscriber)
// 	if err != nil {
// 		t.Fatalf("[after rollback] failed to get `subscriptionsOfSubscriber` err: %v", err)
// 	}
// 	if len(subscriptionsOfSubscriber) != 0 {
// 		t.Fatalf("[after rollback] `subscriptionsOfSubscriber` should be empty, found %v", subscriptionsOfSubscriber)
// 	}
// 	subscriptionsOfSubscription, err = repository.GetSubscriptions(subscription)
// 	if err != nil {
// 		t.Fatalf("[after rollback] failed to get `subscriptionsOfSubscription` err: %v", err)
// 	}
// 	if len(subscriptionsOfSubscription) != 0 {
// 		t.Fatalf("[after rollback] `subscriptionsOfSubscription` be empty, found %v", subscriptionsOfSubscription)
// 	}
// 	subscribersOfSubscriber, err = repository.GetSubscribers(subscriber)
// 	if err != nil {
// 		t.Fatalf("[after rollback] failed to get `subscribersOfSubscriber` err: %v", err)
// 	}
// 	if len(subscribersOfSubscriber) != 0 {
// 		t.Fatalf("[after rollback] `subscribersOfSubscriber` should be empty, found %v", subscribersOfSubscriber)
// 	}
// 	subscribersOfSubscription, err = repository.GetSubscribers(subscription)
// 	if err != nil {
// 		t.Fatalf("[after rollback] failed to get `subscribersOfSubscription` err: %v", err)
// 	}
// 	if len(subscribersOfSubscription) != 0 {
// 		t.Fatalf("[after rollback] `subscribersOfSubscription` should be empty, found %v", subscribersOfSubscription)
// 	}
// }

func TestHasFriendRequestEmpty(t *testing.T) {
	repository := friends.PostgresRepository(database)
	subscriber := randomUid()
	subscription := randomUid()

	exists, err := repository.HasFriendRequest(subscriber, subscription)
	if err != nil {
		t.Fatalf("cannot perform HasFriendRequest err: %v", err)
	}
	if exists {
		t.Fatalf("`exists` should be false")
	}
}

func TestGetSubscribersEmpty(t *testing.T) {
	repository := friends.PostgresRepository(database)
	uid := randomUid()
	subscribers, err := repository.GetSubscribers(uid)
	if err != nil {
		t.Fatalf("failed to perform GetSubscribers err: %v", err)
	}
	if len(subscribers) != 0 {
		t.Fatalf("`subscribers` should be empty")
	}
}

func TestGetSubsriptionsEmpty(t *testing.T) {
	repository := friends.PostgresRepository(database)
	uid := randomUid()
	subscriptions, err := repository.GetSubscriptions(uid)
	if err != nil {
		t.Fatalf("failed to perform GetSubscriptions err: %v", err)
	}
	if len(subscriptions) != 0 {
		t.Fatalf("`subscriptions` should be empty")
	}
}

func TestGetFriendsEmpty(t *testing.T) {
	repository := friends.PostgresRepository(database)
	uid := randomUid()
	friends, err := repository.GetFriends(uid)
	if err != nil {
		t.Fatalf("failed to perform GetFriends err: %v", err)
	}
	if len(friends) != 0 {
		t.Fatalf("`friends` should be empty")
	}
}
