package watchdog

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

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
