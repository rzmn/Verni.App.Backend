package emailSender

import "verni/internal/services/logging"

type Service interface {
	Send(subject string, email string) error
}

type YandexConfig struct {
	Address  string `json:"address"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     string `json:"port"`
}

func YandexService(
	config YandexConfig,
	logger logging.Service,
) Service {
	return &yandexService{
		sender:   config.Address,
		password: config.Password,
		host:     config.Host,
		port:     config.Port,
		logger:   logger,
	}
}
