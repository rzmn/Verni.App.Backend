package watchdog

type Service interface {
	NotifyMessage(message string) error
	NotifyFile(path string) error
}
