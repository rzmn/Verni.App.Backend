package defaultRepository_test

import (
	"encoding/json"
	"io"
	"os"
	"reflect"
	"testing"

	"github.com/rzmn/governi/internal/db"
	postgresDb "github.com/rzmn/governi/internal/db/postgres"
	"github.com/rzmn/governi/internal/repositories/users"
	defaultRepository "github.com/rzmn/governi/internal/repositories/users/default"
	standartOutputLoggingService "github.com/rzmn/governi/internal/services/logging/standartOutput"
	envBasedPathProvider "github.com/rzmn/governi/internal/services/pathProvider/env"

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

func randomUserWithAvatar(hasAvatar bool) users.User {
	var avatarId *users.AvatarId
	if hasAvatar {
		id := users.AvatarId(uuid.New().String())
		avatarId = &id
	} else {
		avatarId = nil
	}
	return users.User{
		Id:          users.UserId(uuid.New().String()),
		DisplayName: uuid.New().String(),
		AvatarId:    avatarId,
	}
}

func TestStore(t *testing.T) {
	storeUser(randomUserWithAvatar(true), t)
	storeUser(randomUserWithAvatar(false), t)
}

func storeUser(user users.User, t *testing.T) {
	repository := defaultRepository.New(database, standartOutputLoggingService.New())

	// if no user with this id, should return []

	shouldBeEmpty, err := repository.GetUsers([]users.UserId{user.Id})
	if err != nil {
		t.Fatalf("failed to get `shouldBeEmpty` err: %v", err)
	}
	if len(shouldBeEmpty) != 0 {
		t.Fatalf("`shouldBeEmpty` is %v, expected empty", shouldBeEmpty)
	}

	// check if store works

	storeTransaction := repository.StoreUser(user)
	if err := storeTransaction.Perform(); err != nil {
		t.Fatalf("failed to perform `storeTransaction` err: %v", err)
	}
	shouldBeSingleUser, err := repository.GetUsers([]users.UserId{user.Id})
	if err != nil {
		t.Fatalf("failed to get `shouldBeSingleUser` err: %v", err)
	}
	if len(shouldBeSingleUser) != 1 || !reflect.DeepEqual(shouldBeSingleUser[0], user) {
		t.Fatalf("`shouldBeSingleUser` is %v, expected %v", shouldBeSingleUser, user)
	}

	// check if another store with same id fails

	userWithSameId := randomUserWithAvatar(true)
	userWithSameId.Id = user.Id
	storeUserWithSameIdAgain := repository.StoreUser(userWithSameId)
	if err := storeUserWithSameIdAgain.Perform(); err == nil {
		t.Fatalf("`storeUserWithSameIdAgain` succeeded, expected to fail")
	}

	// check if rollback works

	if err := storeTransaction.Rollback(); err != nil {
		t.Fatalf("failed to rollback `storeTransaction` err: %v", err)
	}
	shouldBeEmpty, err = repository.GetUsers([]users.UserId{user.Id})
	if err != nil {
		t.Fatalf("[after rollback] failed to get `shouldBeEmpty` err: %v", err)
	}
	if len(shouldBeEmpty) != 0 {
		t.Fatalf("[after rollback] `shouldBeEmpty` is %v, expected empty", shouldBeEmpty)
	}
}

func TestUpdateDisplayName(t *testing.T) {
	repository := defaultRepository.New(database, standartOutputLoggingService.New())
	user := randomUserWithAvatar(true)
	storeTransaction := repository.StoreUser(user)
	if err := storeTransaction.Perform(); err != nil {
		t.Fatalf("failed to perform `storeTransaction` err: %v", err)
	}
	userWithNewDisplayName := user
	userWithNewDisplayName.DisplayName = uuid.New().String()
	updateDisplayNameTransaction := repository.UpdateDisplayName(userWithNewDisplayName.DisplayName, user.Id)
	if err := updateDisplayNameTransaction.Perform(); err != nil {
		t.Fatalf("failed to perform `updateDisplayNameTransaction` err: %v", err)
	}
	shouldBeUserWithNewDisplayName, err := repository.GetUsers([]users.UserId{user.Id})
	if err != nil {
		t.Fatalf("failed to get `shouldBeUserWithNewDisplayName` err: %v", err)
	}
	if len(shouldBeUserWithNewDisplayName) != 1 || !reflect.DeepEqual(shouldBeUserWithNewDisplayName[0], userWithNewDisplayName) {
		t.Fatalf("`shouldBeUserWithNewDisplayName` is %v, expected %v", shouldBeUserWithNewDisplayName, userWithNewDisplayName)
	}
	if err := updateDisplayNameTransaction.Rollback(); err != nil {
		t.Fatalf("failed to rollback `updateDisplayNameTransaction` err: %v", err)
	}
	shouldBeUserWithOldDisplayName, err := repository.GetUsers([]users.UserId{user.Id})
	if err != nil {
		t.Fatalf("failed to get `shouldBeUserWithOldDisplayName` err: %v", err)
	}
	if len(shouldBeUserWithOldDisplayName) != 1 || !reflect.DeepEqual(shouldBeUserWithOldDisplayName[0], user) {
		t.Fatalf("`shouldBeUserWithNewDisplayName` is %v, expected %v", shouldBeUserWithOldDisplayName, user)
	}
}

func TestUpdateAvatar(t *testing.T) {
	testUpdateAvatar(randomUserWithAvatar(true), nil, t)
	testUpdateAvatar(randomUserWithAvatar(false), nil, t)

	avatar1 := uuid.New().String()
	testUpdateAvatar(randomUserWithAvatar(false), (*users.AvatarId)(&avatar1), t)
	avatar2 := uuid.New().String()
	testUpdateAvatar(randomUserWithAvatar(false), (*users.AvatarId)(&avatar2), t)
}

func testUpdateAvatar(user users.User, newAvatar *users.AvatarId, t *testing.T) {
	repository := defaultRepository.New(database, standartOutputLoggingService.New())
	storeTransaction := repository.StoreUser(user)
	if err := storeTransaction.Perform(); err != nil {
		t.Fatalf("failed to perform `storeTransaction` err: %v", err)
	}
	userWithNewAvatar := user
	userWithNewAvatar.AvatarId = newAvatar
	updateAvatarTransaction := repository.UpdateAvatarId(newAvatar, user.Id)
	if err := updateAvatarTransaction.Perform(); err != nil {
		t.Fatalf("failed to perform `updateAvatarTransaction` err: %v", err)
	}
	shouldBeUserWithNewAvatar, err := repository.GetUsers([]users.UserId{user.Id})
	if err != nil {
		t.Fatalf("failed to get `shouldBeUserWithNewAvatar` err: %v", err)
	}
	if len(shouldBeUserWithNewAvatar) != 1 || !reflect.DeepEqual(shouldBeUserWithNewAvatar[0], userWithNewAvatar) {
		t.Fatalf("`shouldBeUserWithNewAvatar` is %v, expected %v", shouldBeUserWithNewAvatar, userWithNewAvatar)
	}
	if err := updateAvatarTransaction.Rollback(); err != nil {
		t.Fatalf("failed to rollback `updateAvatarTransaction` err: %v", err)
	}
	shouldBeUserWithOldAvatar, err := repository.GetUsers([]users.UserId{user.Id})
	if err != nil {
		t.Fatalf("failed to get `shouldBeUserWithOldAvatar` err: %v", err)
	}
	if len(shouldBeUserWithOldAvatar) != 1 || !reflect.DeepEqual(shouldBeUserWithOldAvatar[0], user) {
		t.Fatalf("`shouldBeUserWithOldAvatar` is %v, expected %v", shouldBeUserWithOldAvatar, user)
	}
}

func TestSearchNameContainsSubstring(t *testing.T) {
	repository := defaultRepository.New(database, standartOutputLoggingService.New())
	user := randomUserWithAvatar(true)
	storeTransaction := repository.StoreUser(user)
	if err := storeTransaction.Perform(); err != nil {
		t.Fatalf("failed to perform `storeTransaction` err: %v", err)
	}
	shouldContainUser, err := repository.SearchUsers(user.DisplayName[1:4])
	if err != nil {
		t.Fatalf("failed to get `shouldContainUser` err: %v", err)
	}
	if len(shouldContainUser) != 1 || !reflect.DeepEqual(shouldContainUser[0], user) {
		t.Fatalf("`shouldContainUser` is %v, expected %v", shouldContainUser, user)
	}
}

func TestSearchNameContainsWholeString(t *testing.T) {
	repository := defaultRepository.New(database, standartOutputLoggingService.New())
	user := randomUserWithAvatar(true)
	storeTransaction := repository.StoreUser(user)
	if err := storeTransaction.Perform(); err != nil {
		t.Fatalf("failed to perform `storeTransaction` err: %v", err)
	}
	shouldContainUser, err := repository.SearchUsers(user.DisplayName)
	if err != nil {
		t.Fatalf("failed to get `shouldContainUser` err: %v", err)
	}
	if len(shouldContainUser) != 1 || !reflect.DeepEqual(shouldContainUser[0], user) {
		t.Fatalf("`shouldContainUser` is %v, expected %v", shouldContainUser, user)
	}
}

func TestSearchNameDidNotMatch(t *testing.T) {
	repository := defaultRepository.New(database, standartOutputLoggingService.New())
	user := randomUserWithAvatar(true)
	storeTransaction := repository.StoreUser(user)
	if err := storeTransaction.Perform(); err != nil {
		t.Fatalf("failed to perform `storeTransaction` err: %v", err)
	}
	shouldNotContainUser, err := repository.SearchUsers("#$#$!@#$%)")
	if err != nil {
		t.Fatalf("failed to get `shouldNotContainUser` err: %v", err)
	}
	if len(shouldNotContainUser) != 0 {
		t.Fatalf("`shouldNotContainUser` is %v, expected empty", shouldNotContainUser)
	}
}
