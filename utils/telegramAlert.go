package utils

import (
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func RegisterForAlertTelegram() int64 {
	EnvInit()
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_APITOKEN"))
	if err != nil {
		panic(err)
	}

	bot.Debug = true

	updateConfig := tgbotapi.NewUpdate(0)

	// Tell Telegram we should wait up to 30 seconds on each request for an
	// update. This way we can get information just as quickly as making many
	// frequent requests without having to send nearly as many.
	updateConfig.Timeout = 30

	// Start polling Telegram for updates.
	updates, _ := bot.GetUpdates(updateConfig)

	// Let's go through each update that we're getting from Telegram.
	for _, update := range updates {
		// Telegram can send many types of updates depending on what your Bot
		// is up to. We only want to look at messages for now, so we can
		// discard any other updates.
		if update.Message == nil {
			continue
		}

		return update.Message.Chat.ID
	}
	return -1
}

func SendTelegramMessage(chatId int64, message string) error {
	EnvInit()
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_APITOKEN"))
	if err != nil {
		return err
	}

	msg := tgbotapi.NewMessage(chatId, message)
	_, err = bot.Send(msg)
	return err
}
