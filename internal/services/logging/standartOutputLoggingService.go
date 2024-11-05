package logging

import "log"

type standartOutputLoggingService struct{}

func (c *standartOutputLoggingService) LogInfo(format string, v ...any) {
	log.Printf(format, v...)
}

func (c *standartOutputLoggingService) LogError(format string, v ...any) {
	log.Fatalf(format, v...)
}

func (c *standartOutputLoggingService) LogFatal(format string, v ...any) {
	log.Fatalf(format, v...)
}
