package logging

import "log"

type standartOutputLoggingService struct{}

func (c *standartOutputLoggingService) Log(format string, v ...any) {
	log.Printf(format, v...)
}

func (c *standartOutputLoggingService) Fatalf(format string, v ...any) {
	log.Fatalf(format, v...)
}
