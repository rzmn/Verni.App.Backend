package logging

type Service interface {
	LogInfo(format string, v ...any)
	LogError(format string, v ...any)
	LogFatal(format string, v ...any)
}
