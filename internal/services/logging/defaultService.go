package logging

import "log"

type defaultService struct{}

func (c *defaultService) Log(format string, v ...any) {
	log.Printf(format, v...)
}

func (c *defaultService) Fatalf(format string, v ...any) {
	log.Fatalf(format, v...)
}
