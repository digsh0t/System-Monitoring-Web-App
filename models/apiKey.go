package models

import (
	"errors"

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
