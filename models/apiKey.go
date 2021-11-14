package models

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/wintltr/login-api/database"
)

type ApiKey struct {
	ApiName        string `json:"api_name"`
	ApiToken       string `json:"api_token"`
	TelegramUser   string `json:"api_telegram_user"`
	TelegramChatId int64  `json:"api_telegram_chat_id"`
}

//Insert Telegram Key use for alert bot into DB
func InsertTelegramAPIKeyToDB(apiKey ApiKey) error {
	db := database.ConnectDB()
	defer db.Close()

	encryptedToken := AESEncryptKey(apiKey.ApiToken)
	apiKey.ApiName = "Telegram_bot"
	stmt, err := db.Prepare("INSERT INTO api_keys (ak_api_name, ak_api_token, ak_telegram_user, ak_chat_id) VALUES (?,?,?,?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(apiKey.ApiName, encryptedToken, apiKey.TelegramUser, apiKey.TelegramChatId)
	return err
}

//Get Telegram Token from DB for use or display
func GetTelegramAPIKey() (ApiKey, error) {
	db := database.ConnectDB()
	defer db.Close()

	var apiKey ApiKey
	var encryptedToken string
	row := db.QueryRow(`SELECT ak_api_name,ak_api_token, ak_telegram_user, ak_chat_id FROM api_keys WHERE ak_api_name = "Telegram_bot"`)
	err := row.Scan(&apiKey.ApiName, &encryptedToken, &apiKey.TelegramUser, &apiKey.TelegramChatId)
	if row == nil {
		return apiKey, errors.New("telegram api key doesn't exist")
	}
	if err != nil {
		return apiKey, errors.New("fail to retrieve telegram api key info")
	}

	apiKey.ApiToken, err = AESDecryptKey(encryptedToken)
	if err != nil {
		return apiKey, errors.New("fail to decrypt telegram api key")
	}
	return apiKey, err
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

func SendTelegramMessage(message string) error {
	apiKey, err := GetTelegramAPIKey()
	if err != nil {
		return err
	}

	bot, err := tgbotapi.NewBotAPI(apiKey.ApiToken)
	if err != nil {
		return err
	}

	msg := tgbotapi.NewMessage(apiKey.TelegramChatId, message)

	_, err = bot.Send(msg)
	return err
}

func TestTelegramKey(apiToken string) bool {
	resp, err := http.Get("https://api.telegram.org/bot" + apiToken + "/getMe")
	if err != nil {
		return false
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false
	}
	if strings.Contains(string(body), `"ok":false`) {
		return false
	}
	return true
}

func CheckIfUserHasContactBot(apiToken string, username string) (int64, error) {
	bot, err := tgbotapi.NewBotAPI(apiToken)
	if err != nil {
		return -1, err
	}

	bot.Debug = true

	updateConfig := tgbotapi.NewUpdate(0)

	updates, _ := bot.GetUpdates(updateConfig)
	if len(updates) == 0 {
		return -1, nil
	}
	tmp := updates[len(updates)-1]
	if tmp.Message == nil {
		return -1, errors.New("Please send a message to your Telegram Bot before continuing")
	}
	if strings.Contains(tmp.Message.From.UserName, username) {
		return updates[len(updates)-1].Message.Chat.ID, err
	}
	return -1, err
}

func EditTelegramBotKey(apiKey ApiKey) error {
	err := RemoveTelegramAPIKeyFromDB(apiKey.ApiName)
	if err != nil {
		return err
	}
	err = InsertTelegramAPIKeyToDB(apiKey)
	return err
}
