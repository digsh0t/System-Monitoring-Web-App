package routes

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/bitly/go-simplejson"
	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/utils"
)

func UpdateSSHConnection(w http.ResponseWriter, r *http.Request) {
	// Get user input
	var connectionInfo models.SshConnectionInfo
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, "Error while parsing user's format")
		return
	}
	json.Unmarshal(reqBody, &connectionInfo)

	// Update connectionInfo to DB
	status, err := models.UpdateSSHConnection(connectionInfo)

	// Response to user
	returnJson := simplejson.New()
	returnJson.Set("Status", status)
	returnJson.Set("Error", err)
	var statusCode int
	if err != nil {
		statusCode = http.StatusBadRequest
	} else {
		statusCode = http.StatusOK
	}
	utils.JSON(w, statusCode, returnJson)

}
