package routes

import (
	"net/http"

	"github.com/bitly/go-simplejson"
	"github.com/gorilla/mux"
	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/utils"
)

func TestSSHConnection(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sshConnectionId := vars["id"]
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
