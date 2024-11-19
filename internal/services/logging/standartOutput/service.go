package standartOutputLoggingService

import (
	"log"

	"github.com/rzmn/Verni.App.Backend/internal/services/logging"
)

func New() logging.Service {
	return &standartOutputLoggingService{}
}

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
