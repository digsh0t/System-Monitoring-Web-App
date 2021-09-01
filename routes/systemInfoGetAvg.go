package routes

import (
	"encoding/json"
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

func GetSystemInfoRoute(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	if r.Method == "OPTIONS" {
		//CORS
		// return "OKOK"
		json.NewEncoder(w).Encode("OKOK")
		return
	}

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
	sshConnectionId, err := strconv.Atoi(vars["id"])
	if err != nil {
		returnJson.Set("Status", false)
		returnJson.Set("Error", errors.New("invalid SSH Connection id").Error())
		utils.JSON(w, http.StatusBadRequest, returnJson)
		return
	}

	var systemInfo models.SysInfo
	var eventStatus string
	systemInfo, err = models.GetLatestSysInfo(sshConnectionId, 10)
	if err != nil {
		returnJson.Set("Status", false)
		returnJson.Set("Error", errors.New("fail to get system info").Error())
		utils.JSON(w, http.StatusBadRequest, returnJson)
		eventStatus = "failed"
	} else {
		utils.JSON(w, http.StatusOK, systemInfo)
		eventStatus = "successfully"
	}

	// Write Event Web
	description := "Get Average of system info " + eventStatus
	_, err = event.WriteWebEvent(r, "SystemInfo", description)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("Fail to write event").Error())
		return
	}
}
