package logging

import (
	"context"
	"fmt"
	"os"
	"time"
)

type fileLoggingService struct {
	consoleLogger   Service
	logPathProvider func() *string
	delayedMessages []string
	logger          chan string
}

func (c *fileLoggingService) Log(format string, v ...any) {
	message := prepare(format, v...)
	c.consoleLogger.Log(message)
	c.logger <- message
}

func (c *fileLoggingService) Fatalf(format string, v ...any) {
	message := prepare(format, v...)
	c.consoleLogger.Fatalf(message)
	c.logger <- "[fatal] " + message
	close(c.logger)
}

func prepare(format string, v ...any) string {
	startupTime := time.Now()
	message := fmt.Sprintf(format, v...)
	message = fmt.Sprintf("[%s] %s", startupTime.Format("2006.01.02 15:04:05"), message)
	return message
}

func (c *fileLoggingService) logImpl(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case message := <-c.logger:
			path := c.logPathProvider()
			if path != nil {
				f, err := os.OpenFile(*path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
				if err != nil {
					c.delayedMessages = append(c.delayedMessages, message)
					return
				}
				defer f.Close()
				chunk := ""
				for _, message := range c.delayedMessages {
					chunk += message + "\n"
				}
				chunk += message + "\n"
				if _, err = f.WriteString(chunk); err != nil {
					c.delayedMessages = []string{chunk}
					return
				}
				c.delayedMessages = []string{}
			} else {
				c.delayedMessages = append(c.delayedMessages, message)
			}
		}
	}

}
