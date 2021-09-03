package routes

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/bitly/go-simplejson"
	"github.com/wintltr/login-api/auth"
	"github.com/wintltr/login-api/event"
	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/utils"
)

func AddTelegramBotKey(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	if r.Method == "OPTIONS" {
		//CORS
		// return "OKOK"
		json.NewEncoder(w).Encode("OKOK")
		return
	}

	isAuthorized, err := auth.CheckAuth(r, []string{"admin"})
	if err != nil {
		utils.ERROR(w, http.StatusUnauthorized, errors.New("please login").Error())
		return
	}
	if !isAuthorized {
		utils.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized").Error())
		return
	}

	var apiKey models.ApiKey
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to read request body").Error())
		return
	}

	json.Unmarshal(reqBody, &apiKey)
	isValidKey := models.TestTelegramKey(apiKey.ApiToken)
	if !isValidKey {
		utils.ERROR(w, http.StatusBadRequest, errors.New("your telegram bot api key is not valid, please check again").Error())
		return
	}
	apiKey.TelegramChatId = models.CheckIfUserHasContactBot(apiKey.ApiToken, apiKey.TelegramUser)

	if apiKey.TelegramChatId == -1 {
		utils.ERROR(w, http.StatusBadRequest, errors.New("please send a message to your bot before continuing").Error())
		return
	} else {
		err = models.InsertTelegramAPIKeyToDB(apiKey)
		if err != nil {
			utils.ERROR(w, http.StatusBadRequest, errors.New("fail to insert api key to db").Error())
		}
	}

	// Write Event Web
	description := "Telegram bot key added"
	_, err = event.WriteWebEvent(r, "Bot", description)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("Fail to write bot key event").Error())
		return
	}

	returnJson := simplejson.New()
	returnJson.Set("api_token", apiKey.ApiToken)
	utils.JSON(w, http.StatusOK, returnJson)
}
