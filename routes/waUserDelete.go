package routes

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/bitly/go-simplejson"
	"github.com/gorilla/mux"
	"github.com/wintltr/login-api/auth"
	"github.com/wintltr/login-api/event"
	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/utils"
)

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
	var status string
	if err != nil {
		returnJson.Set("Error", "Fail to delete user")
		status = "failed"
	} else {
		returnJson.Set("Error", err)
		status = "successfully"
	}
	utils.JSON(w, http.StatusBadRequest, returnJson)

	// Write Event Web
	description := "Delete web app user " + username + " " + status
	_, err = event.WriteWebEvent(r, "wauser", description)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("Fail to write task event").Error())
		return
	}

}
