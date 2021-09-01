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

func SSHKeyDeleteRoute(w http.ResponseWriter, r *http.Request) {

	// w.Header().Set("Access-Control-Allow-Origin", "*")
	// w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	if r.Method == "OPTIONS" {
		//CORS
		// return "OKOK"
		json.NewEncoder(w).Encode("OKOK")
		return
	}

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

	returnJson := simplejson.New()
	vars := mux.Vars(r)
	sshKeyId, err := strconv.Atoi(vars["id"])
	if err != nil {
		returnJson.Set("Status", false)
		returnJson.Set("Error", errors.New("invalid SSH key id").Error())
		utils.JSON(w, http.StatusBadRequest, returnJson)
		return
	}

	_, err = models.SSHKeyDelete(sshKeyId)
	var eventStatus string
	if err != nil {
		returnJson.Set("Status", false)
		returnJson.Set("Error", errors.New("please delete ssh connections associated with this SSH key first").Error())
		utils.JSON(w, http.StatusBadRequest, returnJson)
		eventStatus = "failed"
	} else {
		returnJson.Set("Status", true)
		returnJson.Set("Error", nil)
		utils.JSON(w, http.StatusOK, returnJson)
		eventStatus = "successfully"
	}

	// Write Event Web
	description := "Delete sshKey from DB " + eventStatus
	_, err = event.WriteWebEvent(r, "SSHKey", description)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("Fail to write event").Error())
		return
	}

}
