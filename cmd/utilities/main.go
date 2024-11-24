package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	standartOutputLoggingService "github.com/rzmn/governi/internal/services/logging/standartOutput"
	"github.com/rzmn/governi/internal/services/pathProvider"
	envBasedPathProvider "github.com/rzmn/governi/internal/services/pathProvider/env"
)

func main() {
	logger := standartOutputLoggingService.New()
	pathProvider := envBasedPathProvider.New(logger)
	args := os.Args[1:]
	command, err := valueForArg(argNameCommandType, args)
	if err != nil {
		logger.LogFatal("failed to get command type: %v", err)
	}
	switch command {
	case commandNameCreateTables:
		configData, err := getConfigData(args, pathProvider)
		if err != nil {
			logger.LogFatal("failed to get config data: %v", err)
		}
		actions, err := createDatabaseActions(configData, logger)
		if err != nil {
			logger.LogFatal("failed to create database actions err: %v", err)
		}
		actions.setup()
	case commandNameDropTables:
		configData, err := getConfigData(args, pathProvider)
		if err != nil {
			logger.LogFatal("failed to get config data: %v", err)
		}
		actions, err := createDatabaseActions(configData, logger)
		if err != nil {
			logger.LogFatal("failed to create database actions err: %v", err)
		}
		actions.drop()
	}
}

const (
	argNameCommandType   = "--command"
	argNameConfigKeyPath = "--config-key-path"
	argNameConfigPath    = "--config-path"
)

const (
	commandNameCreateTables = "create-tables"
	commandNameDropTables   = "drop-tables"
)

const (
	argNotFoundError = "arg not found"
)

func valueForArg(argName string, args []string) (string, error) {
	for i := 0; i < len(args); i += 2 {
		if argName != args[i] {
			continue
		}
		return args[i+1], nil
	}
	return "", errors.New(argNotFoundError)
}

func getConfigData(args []string, pathProvider pathProvider.Service) ([]byte, error) {
	configPath, err := valueForArg(argNameConfigPath, args)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to get config path err: %v", err)
	}
	configFile, err := os.Open(pathProvider.AbsolutePath(configPath))
	if err != nil {
		return []byte{}, fmt.Errorf("failed to open config file err: %v", err)
	}
	defer configFile.Close()
	configData, err := io.ReadAll(configFile)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to read config file err: %v", err)
	}
	keyPathValue, err := valueForArg(argNameConfigKeyPath, args)
	keyPath := []string{}
	if err != nil {
		if err.Error() == argNotFoundError {
			keyPath = []string{}
		} else {
			return []byte{}, fmt.Errorf("failed to get config key path err: %v", err)
		}
	} else {
		for _, element := range strings.Split(keyPathValue, ".") {
			if len(element) == 0 {
				continue
			}
			keyPath = append(keyPath, element)
		}
	}
	for _, key := range keyPath {
		var config map[string]interface{}
		json.Unmarshal([]byte(configData), &config)
		nested := config[key]
		configData, err = json.Marshal(nested)
		if err != nil {
			return []byte{}, fmt.Errorf("failed to serialize key %s keypath %v data %v err: %v", key, keyPath, config, err)
		}
	}
	return configData, nil
}
