package routes

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/bitly/go-simplejson"
	"github.com/gorilla/mux"
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

func AddWebAppUser(w http.ResponseWriter, r *http.Request) {
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

	result, err := models.AddWebAppUser(user)

	// Return json
	returnJson := simplejson.New()
	var statusCode int
	var status string
	returnJson.Set("Status", result)
	if err != nil {
		returnJson.Set("Error", err.Error())
		statusCode = http.StatusBadRequest
		status = "failed"
	} else {
		returnJson.Set("Error", err)
		statusCode = http.StatusOK
		status = "successfully"
	}
	utils.JSON(w, statusCode, returnJson)

	// Write Event Web
	description := "Add new web app user " + user.Username + " " + status
	_, err = models.WriteWebEvent(r, "wauser", description)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("Fail to write task event").Error())
		return
	}

}

func DeleteWebAppUser(w http.ResponseWriter, r *http.Request) {
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

	// Retrieve Id
	vars := mux.Vars(r)
	waUserId, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.ERROR(w, http.StatusUnauthorized, errors.New("Failed to retrieve Id").Error())
		return
	}

	// Get Username
	username, err := models.GetUsernameFromId(waUserId)
	if err != nil {
		utils.ERROR(w, http.StatusUnauthorized, errors.New("Failed to retrieve username").Error())
		return
	}

	result, err := models.DeleteWepAppUser(waUserId)

	// Return json
	returnJson := simplejson.New()
	returnJson.Set("Status", result)
	var statusCode int
	var status string
	if err != nil {
		returnJson.Set("Error", "Fail to delete user")
		status = "failed"
		statusCode = http.StatusBadRequest
	} else {
		returnJson.Set("Error", err)
		status = "successfully"
		statusCode = http.StatusOK
	}
	utils.JSON(w, statusCode, returnJson)

	// Write Event Web
	description := "Delete web app user " + username + " " + status
	_, err = models.WriteWebEvent(r, "wauser", description)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("Fail to write task event").Error())
		return
	}

}

func ListWebAppUser(w http.ResponseWriter, r *http.Request) {
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

	vars := mux.Vars(r)
	waUserId, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, "failed to list web app user")
		return
	}

	var waUser models.User
	waUser, err = models.ListWepAppUser(waUserId)

	// Return json
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
	} else {
		utils.JSON(w, http.StatusOK, waUser)
	}

}

func ListAllWebAppUser(w http.ResponseWriter, r *http.Request) {
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

	var waUserList []models.User
	waUserList, err = models.ListAllWepAppUser()

	// Return json
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("Fail to get all list web app users").Error())
	} else {
		utils.JSON(w, http.StatusOK, waUserList)
	}

}
