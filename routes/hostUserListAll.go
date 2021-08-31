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

func HostUserListAll(w http.ResponseWriter, r *http.Request) {

	//Authorization
	isAuthorized, err := auth.CheckAuth(r, []string{"admin", "user"})
	if err != nil {
		utils.ERROR(w, http.StatusUnauthorized, errors.New("invalid token").Error())
		return
	}
	if !isAuthorized {
		utils.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized").Error())
		return
	}

	returnJson := simplejson.New()
	vars := mux.Vars(r)
	stringId := vars["id"]
	intId, err := strconv.Atoi(stringId)
	if err != nil {
		returnJson.Set("Status", false)
		returnJson.Set("Error", errors.New("Fail to convert id string to int").Error())
		utils.JSON(w, http.StatusBadRequest, returnJson)
		return
	}
	users, err := models.HostUserListAll(intId)
	if err != nil {
		returnJson.Set("Status", false)
		returnJson.Set("Error", errors.New("Fail to get users from host").Error())
		utils.JSON(w, http.StatusBadRequest, returnJson)
		return
	}

	// Write Event Web
	hostname, err := models.GetSshHostnameFromId(intId)
	if err != nil {
		returnJson.Set("Status", false)
		returnJson.Set("Error", "Fail to get hostname from id")
		utils.JSON(w, http.StatusBadRequest, returnJson)
		return
	}
	id, err := auth.ExtractUserId(r)
	if err != nil {
		returnJson.Set("Status", false)
		returnJson.Set("Error", "Fail to get id of creator")
		utils.JSON(w, http.StatusBadRequest, returnJson)
		return
	}

	var eventWeb event.EventWeb = event.EventWeb{
		EventWebType:        "HostUser",
		EventWebDescription: "List all user from " + hostname,
		EventWebCreatorId:   id,
	}
	_, err = eventWeb.WriteWebEvent()
	if err != nil {
		returnJson.Set("Status", false)
		returnJson.Set("Error", "Fail to write web event")
		utils.JSON(w, http.StatusBadRequest, returnJson)
		return
	}

	// Return Json
	utils.JSON(w, http.StatusOK, users)
	return

}
