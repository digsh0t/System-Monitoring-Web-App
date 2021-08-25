package models

import (
	"errors"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/wintltr/login-api/database"
)

type ApiKey struct {
	ApiName  string `json:"api_name"`
	ApiToken string `json:"api_token"`
}

//Insert Telegram Key use for alert bot into DB
func InsertTelegramAPIKeyToDB(apiKey ApiKey) error {
	db := database.ConnectDB()
	defer db.Close()

	encryptedToken := AESEncryptKey(apiKey.ApiToken)
	apiKey.ApiName = "Telegram_bot"
	stmt, err := db.Prepare("INSERT INTO api_keys (ak_api_name, ak_api_token) VALUES (?,?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(apiKey.ApiName, encryptedToken)
	return err
}

//Get Telegram Token from DB for use or display
func GetTelegramToken() (string, error) {
	db := database.ConnectDB()
	defer db.Close()

	var encryptedToken string
	row := db.QueryRow(`SELECT ak_api_token FROM api_keys WHERE ak_api_name = "Telegram_bot"`)
	err := row.Scan(&encryptedToken)
	if row == nil {
		return "", errors.New("telegram api key doesn't exist")
	}
	if err != nil {
		return "", errors.New("fail to retrieve telegram api key info")
	}

	apiToken, err := AESDecryptKey(encryptedToken)
	if err != nil {
		return "", errors.New("fail to decrypt telegram api key")
	}
	return apiToken, err
}

func RemoveTelegramAPIKeyFromDB(apiName string) error {
	db := database.ConnectDB()
	defer db.Close()

	stmt, err := db.Prepare("DELETE FROM api_keys WHERE ak_api_name = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	res, err := stmt.Exec(apiName)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if rows == 0 {
		return errors.New("no API Keys with this name exist")
	}
	return err
}

func RegisterForAlertTelegram() (int64, error) {
	telegramApiToken, err := GetTelegramToken()
	if err != nil {
		return -1, err
	}
	bot, err := tgbotapi.NewBotAPI(telegramApiToken)
	if err != nil {
		return -1, err
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

		return update.Message.Chat.ID, err
	}
	return -1, err
}

func SendTelegramMessage(chatId int64, message string) error {
	telegramApiToken, err := GetTelegramToken()
	if err != nil {
		return err
	}
	bot, err := tgbotapi.NewBotAPI(telegramApiToken)
	if err != nil {
		return err
	}

	msg := tgbotapi.NewMessage(chatId, message)
	_, err = bot.Send(msg)
	return err
}
