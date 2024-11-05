package pushNotifications_test

import (
	"encoding/json"
	"io"
	"os"
	"testing"
	"verni/internal/db"
	"verni/internal/repositories/pushNotifications"
	"verni/internal/services/logging"
	"verni/internal/services/pathProvider"

	"github.com/google/uuid"
)

var (
	database db.DB
)

func TestMain(m *testing.M) {
	logger := logging.TestService()
	pathProvider := pathProvider.VerniEnvService(logger)
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
		var config db.PostgresConfig
		json.Unmarshal([]byte(configData), &config)
		db, err := db.Postgres(config, logger)
		if err != nil {
			logger.LogFatal("failed to init db err: %v", err)
		}
		return db
	}()
	code := m.Run()

	os.Exit(code)
}

func randomUid() pushNotifications.UserId {
	return pushNotifications.UserId(uuid.New().String())
}

func TestStorePushToken(t *testing.T) {
	repository := pushNotifications.PostgresRepository(database, logging.TestService())

	// initially token should be nil

	uid := randomUid()
	shouldBeEmpty, err := repository.GetPushToken(uid)
	if err != nil {
		t.Fatalf("failed to get `shouldBeEmpty` err: %v", err)
	}
	if shouldBeEmpty != nil {
		t.Fatalf("`shouldBeEmpty` unexpected value: %s", *shouldBeEmpty)
	}

	// test store token when there is no token set previously

	token := uuid.New().String()
	storeTransaction := repository.StorePushToken(uid, token)
	if err := storeTransaction.Perform(); err != nil {
		t.Fatalf("failed to perform `storeTransaction` err: %v", err)
	}
	shouldBeEqualToToken, err := repository.GetPushToken(uid)
	if err != nil {
		t.Fatalf("failed to get `shouldBeEqualToToken` err: %v", err)
	}
	if shouldBeEqualToToken == nil {
		t.Fatalf("`shouldBeEqualToToken` is nil")
	}
	if *shouldBeEqualToToken != token {
		t.Fatalf("`shouldBeEqualToToken` should be equal to %s, found %s", token, *shouldBeEqualToToken)
	}

	// test store token when there is some token set previously

	newToken := uuid.New().String()
	updateTransaction := repository.StorePushToken(uid, newToken)
	if err := updateTransaction.Perform(); err != nil {
		t.Fatalf("failed to perform `updateTransaction` err: %v", err)
	}
	shouldBeEqualToNewToken, err := repository.GetPushToken(uid)
	if err != nil {
		t.Fatalf("failed to get `shouldBeEqualToNewToken` err: %v", err)
	}
	if shouldBeEqualToNewToken == nil {
		t.Fatalf("`shouldBeEqualToNewToken` is nil")
	}
	if *shouldBeEqualToNewToken != newToken {
		t.Fatalf("`shouldBeEqualToNewToken` should be equal to %s, found %s", token, *shouldBeEqualToNewToken)
	}

	// test rollback store token when there is some token set previously

	if err := updateTransaction.Rollback(); err != nil {
		t.Fatalf("failed to rollback `updateTransaction` err: %v", err)
	}
	shouldBeEqualToToken, err = repository.GetPushToken(uid)
	if err != nil {
		t.Fatalf("[after rollback] failed to get `shouldBeEqualToToken` err: %v", err)
	}
	if shouldBeEqualToToken == nil {
		t.Fatalf("[after rollback] `shouldBeEqualToToken` is nil")
	}
	if *shouldBeEqualToToken != token {
		t.Fatalf("[after rollback] `shouldBeEqualToToken` should be equal to %s, found %s", token, *shouldBeEqualToToken)
	}

	// test rollback store token when there is no token set previously

	if err := storeTransaction.Rollback(); err != nil {
		t.Fatalf("failed to rollback `storeTransaction` err: %v", err)
	}
	shouldBeEmpty, err = repository.GetPushToken(uid)
	if err != nil {
		t.Fatalf("[after rollback] failed to get `shouldBeEmpty` err: %v", err)
	}
	if shouldBeEmpty != nil {
		t.Fatalf("[after rollback] `shouldBeEmpty` unexpected value: %s", *shouldBeEmpty)
	}
}
