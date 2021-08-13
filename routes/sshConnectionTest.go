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

func TestSSHConnection(w http.ResponseWriter, r *http.Request) {
	isAuthorized, err := auth.CheckAuth(r, []string{"admin", "user"})
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("invalid token").Error())
		return
	}
	if !isAuthorized {
		utils.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized").Error())
		return
	}

	vars := mux.Vars(r)
	sshConnectionId, _ := strconv.Atoi(vars["id"])
	sshConnection, err := models.GetSSHConnectionFromId(sshConnectionId)

	status := false
	returnJson := simplejson.New()

	if err != nil {
		returnJson.Set("Status", status)
		returnJson.Set("Error", err.Error())
		utils.JSON(w, http.StatusBadRequest, returnJson)
	}

	status, err = sshConnection.TestConnectionPublicKey()
	returnJson.Set("Status", status)
	returnJson.Set("Error", err.Error())
	if err != nil {
		utils.JSON(w, http.StatusBadRequest, returnJson)
	} else {
		utils.JSON(w, http.StatusOK, returnJson)
	}
}
