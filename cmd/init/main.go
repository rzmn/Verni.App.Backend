package main

import (
	"encoding/json"
	"io"
	"os"
	"verni/internal/db"
	"verni/internal/services/logging"
	"verni/internal/services/pathProvider"
)

func main() {
	logger := logging.StandartOutput()
	pathProvider := pathProvider.VerniEnvService(logger)
	configFile, err := os.Open(pathProvider.AbsolutePath("./config/prod/verni.json"))
	if err != nil {
		logger.Fatalf("failed to open config file: %s", err)
	}
	defer configFile.Close()
	configData, err := io.ReadAll(configFile)
	if err != nil {
		logger.Fatalf("failed to read config file: %s", err)
	}
	type Module struct {
		Type   string                 `json:"type"`
		Config map[string]interface{} `json:"config"`
	}
	type Config struct {
		Storage Module `json:"storage"`
	}
	var config Config
	json.Unmarshal([]byte(configData), &config)
	logger.Log("initializing with config %v", config)
	database := func() db.DB {
		switch config.Storage.Type {
		case "postgres":
			data, err := json.Marshal(config.Storage.Config)
			if err != nil {
				logger.Fatalf("failed to serialize ydb config err: %v", err)
			}
			var postgresConfig db.PostgresConfig
			json.Unmarshal(data, &postgresConfig)
			logger.Log("creating postgres with config %v", postgresConfig)
			db, err := db.Postgres(postgresConfig, logger)
			if err != nil {
				logger.Fatalf("failed to initialize postgres err: %v", err)
			}
			logger.Log("initialized postgres")
			return db
		default:
			logger.Fatalf("unknown storage type %s", config.Storage.Type)
			return nil
		}
	}()
	defer database.Close()

	createDatabaseActions(database, logger).setup()
}
