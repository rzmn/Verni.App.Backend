package logging

import "context"

type Service interface {
	Log(format string, v ...any)
	Fatalf(format string, v ...any)
}

func StandartOutput() Service {
	return &standartOutputLoggingService{}
}

func FileLoggerService(logPathProvider func() *string) Service {
	logger := &fileLoggingService{
		consoleLogger:   StandartOutput(),
		delayedMessages: []string{},
		logPathProvider: logPathProvider,
		logger:          make(chan string, 10),
	}
	go logger.logImpl(context.Background())
	return logger
}

func TestService() Service {
	return &standartOutputLoggingService{}
}
