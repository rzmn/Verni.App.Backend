package main

import (
	"encoding/json"
	"io"
	"os"
	"verni/internal/db"
	postgresDb "verni/internal/db/postgres"
	standartOutputLoggingService "verni/internal/services/logging/standartOutput"
	envBasedPathProvider "verni/internal/services/pathProvider/env"
)

func main() {
	logger := standartOutputLoggingService.New()
	pathProvider := envBasedPathProvider.New(logger)
	configFile, err := os.Open(pathProvider.AbsolutePath("./config/prod/verni.json"))
	if err != nil {
		logger.LogFatal("failed to open config file: %s", err)
	}
	defer configFile.Close()
	configData, err := io.ReadAll(configFile)
	if err != nil {
		logger.LogFatal("failed to read config file: %s", err)
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
	logger.LogInfo("initializing with config %v", config)
	database := func() db.DB {
		switch config.Storage.Type {
		case "postgres":
			data, err := json.Marshal(config.Storage.Config)
			if err != nil {
				logger.LogFatal("failed to serialize ydb config err: %v", err)
			}
			var postgresConfig postgresDb.PostgresConfig
			json.Unmarshal(data, &postgresConfig)
			logger.LogInfo("creating postgres with config %v", postgresConfig)
			db, err := postgresDb.Postgres(postgresConfig, logger)
			if err != nil {
				logger.LogFatal("failed to initialize postgres err: %v", err)
			}
			logger.LogInfo("initialized postgres")
			return db
		default:
			logger.LogFatal("unknown storage type %s", config.Storage.Type)
			return nil
		}
	}()
	defer database.Close()

	createDatabaseActions(database, logger).setup()
}
