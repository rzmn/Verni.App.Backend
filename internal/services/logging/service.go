package logging

import (
	"context"
	"sync"
	"verni/internal/services/watchdog"
)

type Service interface {
	LogInfo(format string, v ...any)
	LogError(format string, v ...any)
	LogFatal(format string, v ...any)
}

func StandartOutput() Service {
	return &standartOutputLoggingService{}
}

type ProdLoggerConfig struct {
	Watchdog         watchdog.Service
	LoggingDirectory string
}

func Prod(configProvider func() *ProdLoggerConfig) Service {
	logger := &prodLoggingService{
		consoleLogger: StandartOutput(),
		watchdogProvider: func() *watchdog.Service {
			config := configProvider()
			if config == nil {
				return nil
			}
			return &config.Watchdog
		},
		logsDirectoryProvider: func() *string {
			config := configProvider()
			if config == nil {
				return nil
			}
			return &config.LoggingDirectory
		},
		wg:                           sync.WaitGroup{},
		logger:                       make(chan func(), 10),
		delayedLinesToWriteToLogFile: []string{},
		watchdogContext:              createWatchdogContext(),
		delayedWatchdogCalls:         []func(watchdog.Service){},
	}
	go logger.logImpl(context.Background())
	return logger
}

func TestService() Service {
	return StandartOutput()
}
