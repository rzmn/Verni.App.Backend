package pathProvider

import (
	"os"
	"verni/internal/services/logging"
)

type Service interface {
	AbsolutePath(relative string) string
}

func DefaultService(root string, logger logging.Service) Service {
	logger.Log("override relative paths root: %s", root)
	return &defaultService{
		root: root,
	}
}

func VerniEnvService(logger logging.Service) Service {
	root, present := os.LookupEnv("VERNI_PROJECT_ROOT")
	if !present {
		logger.Fatalf("`VERNI_PROJECT_ROOT` should be set")
	}
	return DefaultService(root, logger)
}
