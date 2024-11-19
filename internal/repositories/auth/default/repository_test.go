package defaultRepository_test

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/rzmn/governi/internal/db"
	postgresDb "github.com/rzmn/governi/internal/db/postgres"
	"github.com/rzmn/governi/internal/repositories/auth"
	defaultRepository "github.com/rzmn/governi/internal/repositories/auth/default"
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

func randomUid() auth.UserId {
	return auth.UserId(uuid.New().String())
}

func randomEmail() string {
	return strings.ReplaceAll(fmt.Sprintf("%s.verni.co", uuid.New().String()), "-", "")
}

func TestGetUserInfo(t *testing.T) {
	repository := defaultRepository.New(database, standartOutputLoggingService.New())
	userId := randomUid()
	userEmail := randomEmail()
	userToken := uuid.New().String()
	userPassword := uuid.New().String()

	userInfo, err := repository.GetUserInfo(userId)
	if err == nil {
		t.Fatalf("[initial] expected to get error from `GetUserInfo` info, found nil")
	}
	createUserTransaction := repository.CreateUser(userId, userEmail, userPassword, userToken)
	if err := createUserTransaction.Perform(); err != nil {
		t.Fatalf("failed to perform `createUserTransaction` err: %v", err)
	}
	userInfo, err = repository.GetUserInfo(userId)
	if err != nil {
		t.Fatalf("failed to get `userInfo` err: %v", err)
	}
	if userInfo.UserId != userId || userInfo.Email != userEmail || userInfo.RefreshToken != userToken || userInfo.EmailVerified {
		t.Fatalf("user info did not match, expected: %v found: %v", auth.UserInfo{
			UserId:        userId,
			Email:         userEmail,
			RefreshToken:  userInfo.RefreshToken,
			EmailVerified: false,
		}, userInfo)
	}
	if err := createUserTransaction.Rollback(); err != nil {
		t.Fatalf("failed to rollback `createUserTransaction` err: %v", err)
	}
	userInfo, err = repository.GetUserInfo(userId)
	if err == nil {
		t.Fatalf("[after rollback] expected to get error from `GetUserInfo` info, found nil")
	}
}

func TestMarkUserEmailValidated(t *testing.T) {
	repository := defaultRepository.New(database, standartOutputLoggingService.New())
	userId := randomUid()
	userEmail := randomEmail()
	userToken := uuid.New().String()
	userPassword := uuid.New().String()

	if err := repository.MarkUserEmailValidated(userId).Perform(); err == nil {
		t.Fatalf("[initial] expected to get error from `MarkUserEmailValidated` info, found nil")
	}
	createUserTransaction := repository.CreateUser(userId, userEmail, userPassword, userToken)
	if err := createUserTransaction.Perform(); err != nil {
		t.Fatalf("failed to perform `createUserTransaction` err: %v", err)
	}
	markValidatedTransaction := repository.MarkUserEmailValidated(userId)
	if err := markValidatedTransaction.Perform(); err != nil {
		t.Fatalf("failed to perform `markValidatedTransaction` err: %v", err)
	}
	userInfo, err := repository.GetUserInfo(userId)
	if err != nil {
		t.Fatalf("failed to get `userInfo` err: %v", err)
	}
	if !userInfo.EmailVerified {
		t.Fatalf("`userInfo.EmailVerified` should be true")
	}
	if err := markValidatedTransaction.Rollback(); err != nil {
		t.Fatalf("failed to rollback `markValidatedTransaction` err: %v", err)
	}
	userInfo, err = repository.GetUserInfo(userId)
	if err != nil {
		t.Fatalf("[after rollback] failed to get `userInfo` err: %v", err)
	}
	if userInfo.EmailVerified {
		t.Fatalf("[after rollback] `userInfo.EmailVerified` should be false")
	}
}

func TestIsUserExists(t *testing.T) {
	repository := defaultRepository.New(database, standartOutputLoggingService.New())
	userId := randomUid()
	userEmail := randomEmail()
	userToken := uuid.New().String()
	userPassword := uuid.New().String()

	userExists, err := repository.IsUserExists(userId)
	if err != nil {
		t.Fatalf("[initial] failed to get `userExists` err: %v", err)
	}
	if userExists {
		t.Fatalf("[initial] `userExists` should be false")
	}
	createUserTransaction := repository.CreateUser(userId, userEmail, userPassword, userToken)
	if err := createUserTransaction.Perform(); err != nil {
		t.Fatalf("failed to perform `createUserTransaction` err: %v", err)
	}
	userExists, err = repository.IsUserExists(userId)
	if err != nil {
		t.Fatalf("failed to get `userExists` err: %v", err)
	}
	if !userExists {
		t.Fatalf("`userExists` should be true")
	}
	if err := createUserTransaction.Rollback(); err != nil {
		t.Fatalf("failed to rollback `createUserTransaction` err: %v", err)
	}
	userExists, err = repository.IsUserExists(userId)
	if err != nil {
		t.Fatalf("[after rollback] failed to get `userExists` err: %v", err)
	}
	if userExists {
		t.Fatalf("[after rollback] `userExists` should be false")
	}
}

func TestCheckCredentials(t *testing.T) {
	repository := defaultRepository.New(database, standartOutputLoggingService.New())
	userId := randomUid()
	userEmail := randomEmail()
	userToken := uuid.New().String()
	userPassword := uuid.New().String()

	_, err := repository.CheckCredentials(userEmail, userPassword)
	if err == nil {
		t.Fatalf("[initial] expected to get error from `CheckCredentials` info, found nil")
	}
	createUserTransaction := repository.CreateUser(userId, userEmail, userPassword, userToken)
	if err := createUserTransaction.Perform(); err != nil {
		t.Fatalf("failed to perform `createUserTransaction` err: %v", err)
	}
	passed, err := repository.CheckCredentials(userEmail, userPassword)
	if err != nil {
		t.Fatalf("failed to run `CheckCredentials` err: %v", err)
	}
	if !passed {
		t.Fatalf("`CheckCredentials` should be true, arg: %v", userPassword)
	}
	wrongPassword := uuid.New().String()
	shouldNotPass, err := repository.CheckCredentials(userEmail, wrongPassword)
	if err != nil {
		t.Fatalf("failed to get `shouldNotPass` err: %v", err)
	}
	if shouldNotPass {
		t.Fatalf("`shouldNotPass` should be false, arg: %v", wrongPassword)
	}
}

func TestGetUserIdByEmail(t *testing.T) {
	repository := defaultRepository.New(database, standartOutputLoggingService.New())
	userId := randomUid()
	userEmail := randomEmail()
	userToken := uuid.New().String()
	userPassword := uuid.New().String()

	shouldBeNil, err := repository.GetUserIdByEmail(userEmail)
	if err != nil {
		t.Fatalf("[initial] failed to get `shouldBeNil` info, found nil")
	}
	if shouldBeNil != nil {
		t.Fatalf("[initial] `shouldBeNil` should be nil, found %s", *shouldBeNil)
	}
	createUserTransaction := repository.CreateUser(userId, userEmail, userPassword, userToken)
	if err := createUserTransaction.Perform(); err != nil {
		t.Fatalf("failed to perform `createUserTransaction` err: %v", err)
	}
	shouldBeEqualToUserId, err := repository.GetUserIdByEmail(userEmail)
	if err != nil {
		t.Fatalf("failed to get `shouldBeEqualToUserId` err: %v", err)
	}
	if shouldBeEqualToUserId == nil {
		t.Fatalf("`shouldBeEqualToUserId` should not be nil")
	}
	if *shouldBeEqualToUserId != userId {
		t.Fatalf("`shouldBeEqualToUserId` should not be equal to %s, found %s", userId, *shouldBeEqualToUserId)
	}
	if err := createUserTransaction.Rollback(); err != nil {
		t.Fatalf("failed to rollback `createUserTransaction` err: %v", err)
	}
	shouldBeNil, err = repository.GetUserIdByEmail(userEmail)
	if err != nil {
		t.Fatalf("[after rollback] failed to get `shouldBeNil` info, found nil")
	}
	if shouldBeNil != nil {
		t.Fatalf("[after rollback] `shouldBeNil` should be nil, found %s", *shouldBeNil)
	}
}

func TestUpdateRefreshToken(t *testing.T) {
	repository := defaultRepository.New(database, standartOutputLoggingService.New())
	userId := randomUid()
	userEmail := randomEmail()
	userToken := uuid.New().String()
	userPassword := uuid.New().String()

	if err := repository.UpdateRefreshToken(userId, uuid.New().String()).Perform(); err == nil {
		t.Fatalf("[initial] expected to get error from `UpdateRefreshToken` info, found nil")
	}
	createUserTransaction := repository.CreateUser(userId, userEmail, userPassword, userToken)
	if err := createUserTransaction.Perform(); err != nil {
		t.Fatalf("failed to perform `createUserTransaction` err: %v", err)
	}
	newRefreshToken := uuid.New().String()
	updateRefreshTokenTransaction := repository.UpdateRefreshToken(userId, newRefreshToken)
	if err := updateRefreshTokenTransaction.Perform(); err != nil {
		t.Fatalf("failed to perform `updateRefreshTokenTransaction` err: %v", err)
	}
	userInfo, err := repository.GetUserInfo(userId)
	if err != nil {
		t.Fatalf("failed to get `userInfo` err: %v", err)
	}
	if userInfo.RefreshToken != newRefreshToken {
		t.Fatalf("userInfo.RefreshToken is not equal to newRefreshToken: %s != %s", userInfo.RefreshToken, newRefreshToken)
	}
	if err := updateRefreshTokenTransaction.Rollback(); err != nil {
		t.Fatalf("failed to rollback `updateRefreshTokenTransaction` err: %v", err)
	}
	userInfo, err = repository.GetUserInfo(userId)
	if err != nil {
		t.Fatalf("[after rollback] failed to get `userInfo` err: %v", err)
	}
	if userInfo.RefreshToken != userToken {
		t.Fatalf("[after rollback] userInfo.RefreshToken is not equal to userToken: %s != %s", userInfo.RefreshToken, userToken)
	}
}

func TestUpdatePassword(t *testing.T) {
	repository := defaultRepository.New(database, standartOutputLoggingService.New())
	userId := randomUid()
	userEmail := randomEmail()
	userToken := uuid.New().String()
	userPassword := uuid.New().String()

	if err := repository.UpdatePassword(userId, uuid.New().String()).Perform(); err == nil {
		t.Fatalf("[initial] expected to get error from `UpdatePassword` info, found nil")
	}
	createUserTransaction := repository.CreateUser(userId, userEmail, userPassword, userToken)
	if err := createUserTransaction.Perform(); err != nil {
		t.Fatalf("failed to perform `createUserTransaction` err: %v", err)
	}
	newPassword := uuid.New().String()
	updatePasswordTransaction := repository.UpdatePassword(userId, newPassword)
	if err := updatePasswordTransaction.Perform(); err != nil {
		t.Fatalf("failed to perform `updatePasswordTransaction` err: %v", err)
	}
	shouldBeTrue, err := repository.CheckCredentials(userEmail, newPassword)
	if err != nil {
		t.Fatalf("failed to perform `CheckCredentials` err: %v", err)
	}
	if !shouldBeTrue {
		t.Fatalf("`CheckCredentials` should return true")
	}
	if err := updatePasswordTransaction.Rollback(); err != nil {
		t.Fatalf("failed to rollback `updateRefreshTokenTransaction` err: %v", err)
	}
	shouldBeTrue, err = repository.CheckCredentials(userEmail, userPassword)
	if err != nil {
		t.Fatalf("[after rollback] failed to perform `CheckCredentials` err: %v", err)
	}
	if !shouldBeTrue {
		t.Fatalf("[after rollback] `CheckCredentials` should return true")
	}
}

func TestUpdateEmail(t *testing.T) {
	repository := defaultRepository.New(database, standartOutputLoggingService.New())
	userId := randomUid()
	userEmail := randomEmail()
	userToken := uuid.New().String()
	userPassword := uuid.New().String()

	if err := repository.UpdateEmail(userId, uuid.New().String()).Perform(); err == nil {
		t.Fatalf("[initial] expected to get error from `UpdateEmail` info, found nil")
	}
	createUserTransaction := repository.CreateUser(userId, userEmail, userPassword, userToken)
	if err := createUserTransaction.Perform(); err != nil {
		t.Fatalf("failed to perform `createUserTransaction` err: %v", err)
	}
	if err := repository.MarkUserEmailValidated(userId).Perform(); err != nil {
		t.Fatalf("[initial] `MarkUserEmailValidated` failed err: %v", err)
	}
	newEmail := randomEmail()
	updateEmailTransaction := repository.UpdateEmail(userId, newEmail)
	if err := updateEmailTransaction.Perform(); err != nil {
		t.Fatalf("failed to perform `updateEmailTransaction` err: %v", err)
	}
	userInfo, err := repository.GetUserInfo(userId)
	if err != nil {
		t.Fatalf("failed to get `userInfo` err: %v", err)
	}
	if userInfo.Email != newEmail || userInfo.EmailVerified {
		t.Fatalf("userInfo should have new and unverified email, found verfified=%t email=%s", userInfo.EmailVerified, userInfo.Email)
	}
	if err := updateEmailTransaction.Rollback(); err != nil {
		t.Fatalf("failed to rollback `updateEmailTransaction` err: %v", err)
	}
	userInfo, err = repository.GetUserInfo(userId)
	if err != nil {
		t.Fatalf("[after rollback] failed to get `userInfo` err: %v", err)
	}
	if userInfo.Email != userEmail || !userInfo.EmailVerified {
		t.Fatalf("[after rollback] userInfo should have old and verified email, found verified=%t email=%s", userInfo.EmailVerified, userInfo.Email)
	}
}
