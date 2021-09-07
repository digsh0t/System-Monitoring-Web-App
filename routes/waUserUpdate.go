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

func UpdateWebAppUser(w http.ResponseWriter, r *http.Request) {
	//Authorization
	isAuthorized, err := auth.CheckAuth(r, []string{"admin"})
	if err != nil {
		utils.ERROR(w, http.StatusUnauthorized, errors.New("invalid token").Error())
		return
	}
	if !isAuthorized {
		utils.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized").Error())
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("Fail to read user data").Error())
		return
	}

	var user models.User
	err = json.Unmarshal(body, &user)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("Fail to parse json format").Error())
		return
	}

	result, err := models.UpdateWepAppUser(user)

	// Return json
	returnJson := simplejson.New()
	returnJson.Set("Status", result)
	var statusCode int
	var status string
	if err != nil {
		returnJson.Set("Error", err.Error())
		status = "failed"
		statusCode = http.StatusBadRequest
	} else {
		returnJson.Set("Error", err)
		status = "successfully"
		statusCode = http.StatusOK
	}
	utils.JSON(w, statusCode, returnJson)

	// Write Event Web
	description := "Update web app user " + user.Username + " " + status
	_, err = models.WriteWebEvent(r, "wauser", description)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("Fail to write task event").Error())
		return
	}

}
