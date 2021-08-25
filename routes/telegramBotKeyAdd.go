package routes

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/bitly/go-simplejson"
	"github.com/wintltr/login-api/auth"
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
	err = models.InsertTelegramAPIKeyToDB(apiKey)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to insert api key to db").Error())
		return
	}
	returnJson := simplejson.New()
	returnJson.Set("api_token", apiKey.ApiToken)
	utils.JSON(w, http.StatusOK, returnJson)
}
