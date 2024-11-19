package defaultRepository_test

import (
	"encoding/json"
	"io"
	"os"
	"reflect"
	"sort"
	"testing"

	"github.com/rzmn/Verni.App.Backend/internal/db"
	postgresDb "github.com/rzmn/Verni.App.Backend/internal/db/postgres"
	"github.com/rzmn/Verni.App.Backend/internal/repositories"
	"github.com/rzmn/Verni.App.Backend/internal/repositories/friends"
	defaultRepository "github.com/rzmn/Verni.App.Backend/internal/repositories/friends/default"
	standartOutputLoggingService "github.com/rzmn/Verni.App.Backend/internal/services/logging/standartOutput"
	envBasedPathProvider "github.com/rzmn/Verni.App.Backend/internal/services/pathProvider/env"

	"github.com/google/uuid"
)

var (
	database db.DB
)

func TestMain(m *testing.M) {
	logger := standartOutputLoggingService.New()
	pathProvider := envBasedPathProvider.New(logger)
	database = func() db.DB {
		configFile, err := os.Open(pathProvider.AbsolutePath("./config/test/postgres_storage.json"))
		if err != nil {
			logger.LogFatal("failed to open config file: %s", err)
		}
		defer configFile.Close()
		configData, err := io.ReadAll(configFile)
		if err != nil {
			logger.LogFatal("failed to read config file: %s", err)
		}
		var config postgresDb.PostgresConfig
		json.Unmarshal([]byte(configData), &config)
		db, err := postgresDb.Postgres(config, logger)
		if err != nil {
			logger.LogFatal("failed to init db err: %v", err)
		}
		return db
	}()
	code := m.Run()

	os.Exit(code)
}

func randomUid() friends.UserId {
	return friends.UserId(uuid.New().String())
}

func TestSubscribtion(t *testing.T) {
	repository := defaultRepository.New(database, standartOutputLoggingService.New())
	subscriber := randomUid()
	subscription := randomUid()

	subsctiptionTransaction := repository.StoreFriendRequest(subscriber, subscription)
	if err := subsctiptionTransaction.Perform(); err != nil {
		t.Fatalf("failed to perform `subsctiptionTransaction` err: %v", err)
	}
	ensureFriendRequest(repository, t, subscriber, subscription, true)
	ensureFriendRequest(repository, t, subscription, subscriber, false)
	ensureSubscriptionsIgnoringOrder(repository, t, subscriber, []friends.UserId{subscription})
	ensureSubscriptionsIgnoringOrder(repository, t, subscription, []friends.UserId{})
	ensureSubscribersIgnoringOrder(repository, t, subscriber, []friends.UserId{})
	ensureSubscribersIgnoringOrder(repository, t, subscription, []friends.UserId{subscriber})
	ensureFriendsIgnoringOrder(repository, t, subscriber, []friends.UserId{})
	ensureFriendsIgnoringOrder(repository, t, subscription, []friends.UserId{})
	if err := subsctiptionTransaction.Rollback(); err != nil {
		t.Fatalf("failed to rollback `subsctiptionTransaction` err: %v", err)
	}
	ensureFriendRequest(repository, t, subscriber, subscription, false)
	ensureFriendRequest(repository, t, subscription, subscriber, false)
	ensureSubscriptionsIgnoringOrder(repository, t, subscriber, []friends.UserId{})
	ensureSubscriptionsIgnoringOrder(repository, t, subscription, []friends.UserId{})
	ensureSubscribersIgnoringOrder(repository, t, subscriber, []friends.UserId{})
	ensureSubscribersIgnoringOrder(repository, t, subscription, []friends.UserId{})
	ensureFriendsIgnoringOrder(repository, t, subscriber, []friends.UserId{})
	ensureFriendsIgnoringOrder(repository, t, subscription, []friends.UserId{})
}

func TestStoreAndRemoveFriendRequest(t *testing.T) {
	repository := defaultRepository.New(database, standartOutputLoggingService.New())
	subscriber := randomUid()
	subscription := randomUid()

	subsctiptionTransaction := repository.StoreFriendRequest(subscriber, subscription)
	if err := subsctiptionTransaction.Perform(); err != nil {
		t.Fatalf("failed to perform `subsctiptionTransaction` err: %v", err)
	}
	ensureFriendRequest(repository, t, subscriber, subscription, true)
	ensureFriendRequest(repository, t, subscription, subscriber, false)
	ensureSubscriptionsIgnoringOrder(repository, t, subscriber, []friends.UserId{subscription})
	ensureSubscriptionsIgnoringOrder(repository, t, subscription, []friends.UserId{})
	ensureSubscribersIgnoringOrder(repository, t, subscriber, []friends.UserId{})
	ensureSubscribersIgnoringOrder(repository, t, subscription, []friends.UserId{subscriber})
	ensureFriendsIgnoringOrder(repository, t, subscriber, []friends.UserId{})
	ensureFriendsIgnoringOrder(repository, t, subscription, []friends.UserId{})
	removeTransaction := repository.RemoveFriendRequest(subscriber, subscription)
	if err := removeTransaction.Perform(); err != nil {
		t.Fatalf("failed to perform `removeTransaction` err: %v", err)
	}
	ensureFriendRequest(repository, t, subscriber, subscription, false)
	ensureFriendRequest(repository, t, subscription, subscriber, false)
	ensureSubscriptionsIgnoringOrder(repository, t, subscriber, []friends.UserId{})
	ensureSubscriptionsIgnoringOrder(repository, t, subscription, []friends.UserId{})
	ensureSubscribersIgnoringOrder(repository, t, subscriber, []friends.UserId{})
	ensureSubscribersIgnoringOrder(repository, t, subscription, []friends.UserId{})
	ensureFriendsIgnoringOrder(repository, t, subscriber, []friends.UserId{})
	ensureFriendsIgnoringOrder(repository, t, subscription, []friends.UserId{})
}

func TestFriendship(t *testing.T) {
	repository := defaultRepository.New(database, standartOutputLoggingService.New())
	subscriber := randomUid()
	subscription := randomUid()

	friendshipTransaction := []repositories.MutationWorkItem{
		repository.StoreFriendRequest(subscriber, subscription),
		repository.StoreFriendRequest(subscription, subscriber),
	}
	for i := 0; i < len(friendshipTransaction); i++ {
		if err := friendshipTransaction[i].Perform(); err != nil {
			t.Fatalf("failed to perform `friendshipTransaction[%d]` err: %v", i, err)
		}
	}
	ensureFriendRequest(repository, t, subscriber, subscription, false)
	ensureFriendRequest(repository, t, subscription, subscriber, false)
	ensureSubscriptionsIgnoringOrder(repository, t, subscriber, []friends.UserId{})
	ensureSubscriptionsIgnoringOrder(repository, t, subscription, []friends.UserId{})
	ensureSubscribersIgnoringOrder(repository, t, subscriber, []friends.UserId{})
	ensureSubscribersIgnoringOrder(repository, t, subscription, []friends.UserId{})
	ensureFriendsIgnoringOrder(repository, t, subscriber, []friends.UserId{subscription})
	ensureFriendsIgnoringOrder(repository, t, subscription, []friends.UserId{subscriber})
	for i := 0; i < len(friendshipTransaction); i++ {
		if err := friendshipTransaction[i].Rollback(); err != nil {
			t.Fatalf("failed to rollback `friendshipTransaction[%d]` err: %v", i, err)
		}
	}
	ensureFriendRequest(repository, t, subscriber, subscription, false)
	ensureFriendRequest(repository, t, subscription, subscriber, false)
	ensureSubscriptionsIgnoringOrder(repository, t, subscriber, []friends.UserId{})
	ensureSubscriptionsIgnoringOrder(repository, t, subscription, []friends.UserId{})
	ensureSubscribersIgnoringOrder(repository, t, subscriber, []friends.UserId{})
	ensureSubscribersIgnoringOrder(repository, t, subscription, []friends.UserId{})
	ensureFriendsIgnoringOrder(repository, t, subscriber, []friends.UserId{})
	ensureFriendsIgnoringOrder(repository, t, subscription, []friends.UserId{})
}

func TestHasFriendRequestEmpty(t *testing.T) {
	repository := defaultRepository.New(database, standartOutputLoggingService.New())
	subscriber := randomUid()
	subscription := randomUid()
	ensureFriendRequest(repository, t, subscriber, subscription, false)
}

func TestGetSubscribersEmpty(t *testing.T) {
	repository := defaultRepository.New(database, standartOutputLoggingService.New())
	uid := randomUid()
	ensureSubscribersIgnoringOrder(repository, t, uid, []friends.UserId{})
}

func TestGetSubsriptionsEmpty(t *testing.T) {
	repository := defaultRepository.New(database, standartOutputLoggingService.New())
	uid := randomUid()
	ensureSubscriptionsIgnoringOrder(repository, t, uid, []friends.UserId{})
}

func TestGetFriendsEmpty(t *testing.T) {
	repository := defaultRepository.New(database, standartOutputLoggingService.New())
	uid := randomUid()
	ensureFriendsIgnoringOrder(repository, t, uid, []friends.UserId{})
}

//

func ensureFriendRequest(repository friends.Repository, t *testing.T, from friends.UserId, to friends.UserId, expectation bool) {
	has, err := repository.HasFriendRequest(from, to)
	if err != nil {
		t.Fatalf("failed to perform `HasFriendRequest` err: %v", err)
	}
	if has != expectation {
		t.Fatalf("`HasFriendRequest` result should be %t, found: %t", expectation, has)
	}
}

func ensureSubscriptionsIgnoringOrder(repository friends.Repository, t *testing.T, userId friends.UserId, expectation []friends.UserId) {
	subscriptions, err := repository.GetSubscriptions(userId)
	if err != nil {
		t.Fatalf("failed to perform `GetSubscriptions` err: %v", err)
	}
	ensureEqualIgnoringOrder(t, subscriptions, expectation, "GetSubscriptions")
}

func ensureSubscribersIgnoringOrder(repository friends.Repository, t *testing.T, userId friends.UserId, expectation []friends.UserId) {
	subscribers, err := repository.GetSubscribers(userId)
	if err != nil {
		t.Fatalf("failed to perform `GetSubscribers` err: %v", err)
	}
	ensureEqualIgnoringOrder(t, subscribers, expectation, "GetSubscribers")
}

func ensureFriendsIgnoringOrder(repository friends.Repository, t *testing.T, userId friends.UserId, expectation []friends.UserId) {
	subscribers, err := repository.GetFriends(userId)
	if err != nil {
		t.Fatalf("failed to get `GetFriends` err: %v", err)
	}
	ensureEqualIgnoringOrder(t, subscribers, expectation, "GetFriends")
}

func ensureEqualIgnoringOrder(t *testing.T, lhs []friends.UserId, rhs []friends.UserId, label string) {
	comparator := func(array []friends.UserId) func(i, j int) bool {
		return func(i, j int) bool {
			return array[i] < array[j]
		}
	}
	sort.Slice(lhs, comparator(lhs))
	sort.Slice(rhs, comparator(rhs))
	if !reflect.DeepEqual(lhs, rhs) {
		t.Fatalf("%s: result should be %v, found %v", label, lhs, rhs)
	}
}
