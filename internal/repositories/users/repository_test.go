package users_test

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"reflect"
	"testing"
	"verni/internal/common"
	"verni/internal/db"
	"verni/internal/repositories/users"

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
	repository := users.PostgresRepository(database)

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
