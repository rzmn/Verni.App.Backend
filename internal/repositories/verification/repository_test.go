package verification_test

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"verni/internal/db"
	"verni/internal/repositories/verification"
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
			logger.Fatalf("failed to open config file: %s", err)
		}
		defer configFile.Close()
		configData, err := io.ReadAll(configFile)
		if err != nil {
			logger.Fatalf("failed to read config file: %s", err)
		}
		var config db.PostgresConfig
		json.Unmarshal([]byte(configData), &config)
		db, err := db.Postgres(config, logger)
		if err != nil {
			logger.Fatalf("failed to init db err: %v", err)
		}
		return db
	}()
	code := m.Run()

	os.Exit(code)
}

func randomEmail() string {
	return strings.ReplaceAll(fmt.Sprintf("%s.verni.co", uuid.New().String()), "-", "")
}

func randomCode() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")[0:6]
}

func TestStore(t *testing.T) {
	repository := verification.PostgresRepository(database, logging.TestService())
	email := randomEmail()

	// if no code was stored, should return nil

	shouldBeEmpty, err := repository.GetEmailVerificationCode(email)
	if err != nil {
		t.Fatalf("failed to get `shouldBeEmpty` err: %v", err)
	}
	if shouldBeEmpty != nil {
		t.Fatalf("`shouldBeEmpty` is %s, expected empty", *shouldBeEmpty)
	}

	// check if store works when there is no token stored previously

	codeToStoreForTheFirstTime := randomCode()
	storeForTheFirstTimeTransaction := repository.StoreEmailVerificationCode(email, codeToStoreForTheFirstTime)
	if err := storeForTheFirstTimeTransaction.Perform(); err != nil {
		t.Fatalf("failed to perform `storeForTheFirstTimeTransaction` err: %v", err)
	}
	shouldBeEqualToCodeStoredAtFirstTime, err := repository.GetEmailVerificationCode(email)
	if err != nil {
		t.Fatalf("failed to get `shouldBeEqualToCodeStoredAtFirstTime` err: %v", err)
	}
	if shouldBeEqualToCodeStoredAtFirstTime == nil {
		t.Fatalf("found empty `shouldBeEqualToCodeStoredAtFirstTime`, expected %s", codeToStoreForTheFirstTime)
	}
	if *shouldBeEqualToCodeStoredAtFirstTime != codeToStoreForTheFirstTime {
		t.Fatalf("codes did not match (%s != %s)", *shouldBeEqualToCodeStoredAtFirstTime, codeToStoreForTheFirstTime)
	}

	// check if store works when there is some token stored previously

	codeToStoreForTheSecondTime := randomCode()
	storeForTheSecondTimeTransaction := repository.StoreEmailVerificationCode(email, codeToStoreForTheSecondTime)
	if err := storeForTheSecondTimeTransaction.Perform(); err != nil {
		t.Fatalf("failed to perform `storeForTheSecondTimeTransaction` err: %v", err)
	}
	shouldBeEqualToCodeStoredAtSecondTime, err := repository.GetEmailVerificationCode(email)
	if err != nil {
		t.Fatalf("failed to get `shouldBeEqualToCodeStoredAtSecondTime` err: %v", err)
	}
	if shouldBeEqualToCodeStoredAtSecondTime == nil {
		t.Fatalf("found empty `shouldBeEqualToCodeStoredAtSecondTime`, expected %s", codeToStoreForTheSecondTime)
	}
	if *shouldBeEqualToCodeStoredAtSecondTime != codeToStoreForTheSecondTime {
		t.Fatalf("codes did not match (%s != %s)", *shouldBeEqualToCodeStoredAtSecondTime, codeToStoreForTheSecondTime)
	}

	// check if rollback works for store when there is some token stored previously

	if err := storeForTheSecondTimeTransaction.Rollback(); err != nil {
		t.Fatalf("failed to rollback `storeForTheSecondTimeTransaction` err: %v", err)
	}
	shouldBeEqualToCodeStoredAtFirstTime, err = repository.GetEmailVerificationCode(email)
	if err != nil {
		t.Fatalf("[after rollback] failed to get `shouldBeEqualToCodeStoredAtFirstTime` err: %v", err)
	}
	if shouldBeEqualToCodeStoredAtFirstTime == nil {
		t.Fatalf("[after rollback] found empty `shouldBeEqualToCodeStoredAtFirstTime`, expected %s", codeToStoreForTheFirstTime)
	}
	if *shouldBeEqualToCodeStoredAtFirstTime != codeToStoreForTheFirstTime {
		t.Fatalf("[after rollback] codes did not match (%s != %s)", *shouldBeEqualToCodeStoredAtFirstTime, codeToStoreForTheFirstTime)
	}

	// check if rollback works for store when there is no token stored previously

	if err := storeForTheFirstTimeTransaction.Rollback(); err != nil {
		t.Fatalf("failed to rollback `storeForTheFirstTimeTransaction` err: %v", err)
	}
	shouldBeEmpty, err = repository.GetEmailVerificationCode(email)
	if err != nil {
		t.Fatalf("[after rollback] failed to get `shouldBeEmpty` err: %v", err)
	}
	if shouldBeEmpty != nil {
		t.Fatalf("`shouldBeEmpty` is %s, expected empty", *shouldBeEmpty)
	}
}

func TestRemoveEmpty(t *testing.T) {
	repository := verification.PostgresRepository(database, logging.TestService())
	email := randomEmail()

	// check if remove works when there is no token stored previously

	removeForTheFirstTime := repository.RemoveEmailVerificationCode(email)
	if err := removeForTheFirstTime.Perform(); err != nil {
		t.Fatalf("failed to perform `removeForTheFirstTime` err: %v", err)
	}
	shouldBeEmpty, err := repository.GetEmailVerificationCode(email)
	if err != nil {
		t.Fatalf("failed to get `shouldBeEmpty` err: %v", err)
	}
	if shouldBeEmpty != nil {
		t.Fatalf("`shouldBeEmpty` is %s, expected empty", *shouldBeEmpty)
	}

	// check if rollback works for remove when there is no token stored previously

	if err := removeForTheFirstTime.Rollback(); err != nil {
		t.Fatalf("failed to rollback `removeForTheFirstTime` err: %v", err)
	}
	shouldBeEmpty, err = repository.GetEmailVerificationCode(email)
	if err != nil {
		t.Fatalf("[after rollback] failed to get `shouldBeEmpty` err: %v", err)
	}
	if shouldBeEmpty != nil {
		t.Fatalf("`shouldBeEmpty` is %s, expected empty", *shouldBeEmpty)
	}
}

func TestRemoveNonEmpty(t *testing.T) {
	repository := verification.PostgresRepository(database, logging.TestService())
	email := randomEmail()
	codeToStore := randomCode()
	storeTransaction := repository.StoreEmailVerificationCode(email, codeToStore)
	if err := storeTransaction.Perform(); err != nil {
		t.Fatalf("failed to perform `storeTransaction` err: %v", err)
	}

	// check if remove works when there is some token stored previously

	removeTransaction := repository.RemoveEmailVerificationCode(email)
	if err := removeTransaction.Perform(); err != nil {
		t.Fatalf("failed to perform `removeTransaction` err: %v", err)
	}
	shouldBeEmpty, err := repository.GetEmailVerificationCode(email)
	if err != nil {
		t.Fatalf("failed to get `shouldBeEmpty` err: %v", err)
	}
	if shouldBeEmpty != nil {
		t.Fatalf("`shouldBeEmpty` is %s, expected empty", *shouldBeEmpty)
	}

	// check if rollback works for remove when there is some token stored previously

	if err := removeTransaction.Rollback(); err != nil {
		t.Fatalf("failed to rollback `removeTransaction` err: %v", err)
	}
	shouldBeEqualToStoredCode, err := repository.GetEmailVerificationCode(email)
	if err != nil {
		t.Fatalf("failed to get `shouldBeEqualToStoredCode` err: %v", err)
	}
	if shouldBeEqualToStoredCode == nil {
		t.Fatalf("found empty `shouldBeEqualToStoredCode`, expected %s", codeToStore)
	}
	if *shouldBeEqualToStoredCode != codeToStore {
		t.Fatalf("codes did not match (%s != %s)", *shouldBeEqualToStoredCode, codeToStore)
	}
}
