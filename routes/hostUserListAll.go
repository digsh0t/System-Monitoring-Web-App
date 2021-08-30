package routes

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/bitly/go-simplejson"
	"github.com/gorilla/mux"
	"github.com/wintltr/login-api/auth"
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
	stringInt, err := strconv.Atoi(stringId)
	if err != nil {
		returnJson.Set("Status", false)
		returnJson.Set("Error", errors.New("Fail to convert id string to int").Error())
		utils.JSON(w, http.StatusBadRequest, returnJson)
		return
	}
	users, err := models.HostUserListAll(stringInt)
	if err != nil {
		returnJson.Set("Status", false)
		returnJson.Set("Error", errors.New("Fail to get users from host").Error())
		utils.JSON(w, http.StatusBadRequest, returnJson)
		return
	}

	// Return Json
	utils.JSON(w, http.StatusOK, users)
	return

}
