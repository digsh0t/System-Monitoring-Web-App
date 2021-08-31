package routes

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/bitly/go-simplejson"
	"github.com/wintltr/login-api/auth"
	"github.com/wintltr/login-api/event"
	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/utils"
)

func GetAllSSHKey(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	if r.Method == "OPTIONS" {
		//CORS
		// return "OKOK"
		json.NewEncoder(w).Encode("OKOK")
		return
	}

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

	sshKeyList, err := models.GetAllSSHKeyFromDB()
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
	}

	// Write Event Web
	returnJson := simplejson.New()
	description := "Get all sshKey"
	_, err = event.WriteWebEvent(r, "SSHKey", description)
	if err != nil {
		returnJson.Set("Status", false)
		returnJson.Set("Error", "Fail to write web event")
		utils.JSON(w, http.StatusBadRequest, returnJson)
		return
	}
	utils.JSON(w, http.StatusOK, sshKeyList)
}
