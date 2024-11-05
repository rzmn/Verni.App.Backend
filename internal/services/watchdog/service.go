package watchdog

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Service interface {
	NotifyMessage(message string) error
	NotifyFile(path string) error
}

type TelegramConfig struct {
	Token     string `json:"token"`
	ChannelId int64  `json:"channelId"`
}

func TelegramService(config TelegramConfig) (Service, error) {
	api, err := tgbotapi.NewBotAPI(config.Token)
	if err != nil {
		return nil, err
	}
	return &telegramService{
		api:       api,
		channelId: config.ChannelId,
	}, nil
}
