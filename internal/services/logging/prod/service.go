package prodLoggingService

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/rzmn/Verni.App.Backend/internal/services/logging"
	standartOutputLoggingService "github.com/rzmn/Verni.App.Backend/internal/services/logging/standartOutput"
	"github.com/rzmn/Verni.App.Backend/internal/services/watchdog"
)

type ProdLoggerConfig struct {
	Watchdog         watchdog.Service
	LoggingDirectory string
}

func New(configProvider func() *ProdLoggerConfig) logging.Service {
	logger := &prodLoggingService{
		consoleLogger: standartOutputLoggingService.New(),
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

type watchdogContext struct {
	capacity int
	size     int
	head     int
	data     []string
}

func (c *watchdogContext) Append(value string) {
	if c.size < c.capacity {
		c.size += 1
	}
	c.data[c.head] = value
	c.head += 1
	c.head %= c.capacity
}

func (c *watchdogContext) Array() []string {
	result := make([]string, c.size)
	startIndex := (c.head - c.size + c.capacity) % c.capacity
	for i := 0; i < c.size; i++ {
		result[i] = c.data[(startIndex+i)%c.capacity]
	}
	return result
}

func createWatchdogContext() watchdogContext {
	capacity := 1000
	return watchdogContext{
		capacity: capacity,
		size:     0,
		head:     0,
		data:     make([]string, capacity),
	}
}

type prodLoggingService struct {
	consoleLogger                logging.Service
	watchdogProvider             func() *watchdog.Service
	logsDirectoryProvider        func() *string
	wg                           sync.WaitGroup
	logger                       chan func()
	delayedLinesToWriteToLogFile []string
	watchdogContext              watchdogContext
	delayedWatchdogCalls         []func(watchdog.Service)
}

func (c *prodLoggingService) LogInfo(format string, v ...any) {
	message := prepare(format, v...)
	c.consoleLogger.LogInfo(message)
	c.wg.Add(1)
	c.logger <- func() {
		c.watchdogContext.Append(message)
		c.writeToFile(message)
	}
}

func (c *prodLoggingService) LogError(format string, v ...any) {
	message := prepare(format, v...)
	c.consoleLogger.LogError(message)
	c.wg.Add(2)
	c.logger <- func() {
		c.watchdogContext.Append(message)
		c.writeToFile("[error] " + message)
	}
	c.logger <- func() {
		c.fireWatchdog("[error] " + message)
	}
}

func (c *prodLoggingService) LogFatal(format string, v ...any) {
	message := prepare(format, v...)
	c.wg.Add(2)
	c.logger <- func() {
		c.watchdogContext.Append(message)
		c.writeToFile("[fatal] " + message)
	}
	c.logger <- func() {
		c.fireWatchdog("[fatal] " + message)
	}
	c.wg.Wait()
	close(c.logger)
	c.consoleLogger.LogFatal(message)
}

func prepare(format string, v ...any) string {
	startupTime := time.Now()
	message := fmt.Sprintf(format, v...)
	message = fmt.Sprintf("[%s] %s", startupTime.Format("2006.01.02 15:04:05"), message)
	return message
}

func (c *prodLoggingService) writeToFile(message string) {
	logsDirectory := c.logsDirectoryProvider()
	if logsDirectory != nil {
		path := getLogPath(*logsDirectory)
		f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			c.delayedLinesToWriteToLogFile = append(c.delayedLinesToWriteToLogFile, message)
			return
		}
		defer f.Close()
		chunk := ""
		for _, message := range c.delayedLinesToWriteToLogFile {
			chunk += message + "\n"
		}
		chunk += message + "\n"
		if _, err = f.WriteString(chunk); err != nil {
			c.delayedLinesToWriteToLogFile = []string{chunk}
			return
		}
		c.delayedLinesToWriteToLogFile = []string{}
	} else {
		c.delayedLinesToWriteToLogFile = append(c.delayedLinesToWriteToLogFile, message)
	}
}

func (c *prodLoggingService) fireWatchdog(message string) {
	file, err := os.CreateTemp("", "watchdogContext")
	if err != nil {
		c.fireWatchdogImpl(func(watchdog watchdog.Service) {
			watchdog.NotifyMessage(fmt.Sprintf("[panic] shutting down wd, reason: cannot create logs file err: %v", err))
			c.watchdogProvider = nil
		})
		return
	}
	context := strings.Join(c.watchdogContext.Array(), "\n")
	if _, err = file.WriteString(context); err != nil {
		c.fireWatchdogImpl(func(watchdog watchdog.Service) {
			watchdog.NotifyMessage(fmt.Sprintf("[panic] shutting down wd, reason: cannot write logs to file err: %v", err))
			c.watchdogProvider = nil
		})
		return
	}
	c.fireWatchdogImpl(func(watchdog watchdog.Service) {
		watchdog.NotifyMessage(fmt.Sprintf("internal error: %s", message))
		watchdog.NotifyFile(file.Name())
	})
}

func (c *prodLoggingService) fireWatchdogImpl(routine func(watchdog watchdog.Service)) {
	c.delayedWatchdogCalls = append(c.delayedWatchdogCalls, routine)
	for _, routine := range c.delayedWatchdogCalls {
		provider := c.watchdogProvider
		if provider == nil {
			c.delayedWatchdogCalls = []func(watchdog.Service){}
			return
		}
		watchdog := provider()
		if watchdog == nil {
			return
		}
		routine(*watchdog)
	}
	c.delayedWatchdogCalls = []func(watchdog.Service){}
}

func getLogPath(directory string) string {
	return filepath.Join(directory, "./1.log")
}

func (c *prodLoggingService) logImpl(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case routine := <-c.logger:
			if routine != nil {
				routine()
			}
			c.wg.Done()
		}
	}
}
