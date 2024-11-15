package defaultRepository_test

import (
	"encoding/json"
	"io"
	"os"
	"testing"
	"verni/internal/db"
	postgresDb "verni/internal/db/postgres"
	"verni/internal/repositories/images"
	defaultRepository "verni/internal/repositories/images/default"
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

func TestUpload(t *testing.T) {
	repository := defaultRepository.New(database, logging.TestService())
	base64 := uuid.New().String()

	transaction := repository.UploadImageBase64(base64)
	uploadedId, err := transaction.Perform()
	if err != nil {
		t.Fatalf("failed to perform transaction err: %v", err)
	}
	shouldContainUploadedId, err := repository.GetImagesBase64([]images.ImageId{uploadedId})
	if err != nil {
		t.Fatalf("failed to get `shouldContainUploadedId` err: %v", err)
	}
	if len(shouldContainUploadedId) != 1 || shouldContainUploadedId[0].Id != uploadedId || shouldContainUploadedId[0].Base64 != base64 {
		t.Fatalf("`shouldContainUploadedId` is %v, expected to contain %s id %s data", shouldContainUploadedId, uploadedId, base64)
	}
	if err := transaction.Rollback(); err != nil {
		t.Fatalf("failed to rollback transaction err: %v", err)
	}
	shouldBeEmpty, err := repository.GetImagesBase64([]images.ImageId{uploadedId})
	if err != nil {
		t.Fatalf("failed to get `shouldBeEmpty` err: %v", err)
	}
	if len(shouldBeEmpty) != 0 {
		t.Fatalf("`shouldBeEmpty` should be empty, found %v", shouldBeEmpty)
	}
}

func TestGetEmpty(t *testing.T) {
	repository := defaultRepository.New(database, logging.TestService())
	id := images.ImageId(uuid.New().String())

	shouldBeEmpty, err := repository.GetImagesBase64([]images.ImageId{id})
	if err != nil {
		t.Fatalf("failed to get `shouldBeEmpty` err: %v", err)
	}
	if len(shouldBeEmpty) != 0 {
		t.Fatalf("`shouldBeEmpty` should be empty, found %v", shouldBeEmpty)
	}
}
