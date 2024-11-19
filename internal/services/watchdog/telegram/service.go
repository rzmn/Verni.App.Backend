package telegramWatchdog

import (
	"github.com/rzmn/governi/internal/services/watchdog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TelegramConfig struct {
	Token     string `json:"token"`
	ChannelId int64  `json:"channelId"`
}

func New(config TelegramConfig) (watchdog.Service, error) {
	api, err := tgbotapi.NewBotAPI(config.Token)
	if err != nil {
		return nil, err
	}
	return &telegramService{
		api:       api,
		channelId: config.ChannelId,
	}, nil
}

type telegramService struct {
	api       *tgbotapi.BotAPI
	channelId int64
}

func (c *telegramService) NotifyMessage(message string) error {
	_, err := c.api.Send(tgbotapi.NewMessage(c.channelId, message))
	return err
}

func (c *telegramService) NotifyFile(path string) error {
	_, err := c.api.Send(tgbotapi.NewDocument(c.channelId, tgbotapi.FilePath(path)))
	return err
}
