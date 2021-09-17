package routes

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/bitly/go-simplejson"
	"github.com/wintltr/login-api/auth"
	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/utils"
)

func HostUserRemove(w http.ResponseWriter, r *http.Request) {

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

	// Retrieve Json Format
	reqBody, err := ioutil.ReadAll(r.Body)

	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("Fail to retrieve json format").Error())
		return
	}
	var hostUser models.HostUserInfo
	json.Unmarshal(reqBody, &hostUser)

	// Remove Host User
	output, err := hostUser.HostUserRemove()

	// Processing Output From Ansible
	status, fatalList, err := models.ProcessingAnsibleOutput(output)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, "fail to process ansible output")
		return
	}

	// Return Json
	returnJson := simplejson.New()
	returnJson.Set("Status", status)
	returnJson.Set("Fatal", fatalList)
	utils.JSON(w, http.StatusBadRequest, returnJson)

	// Write Event Web
	host, err := models.ConvertListIdToHostname(hostUser.SshConnectionIdList)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("Fail to convert id string to hostname").Error())
		return
	}

	description := "User \"" + hostUser.HostUserName + "\" removed from " + host + " successfully"
	_, err = models.WriteWebEvent(r, "HostUser", description)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("Fail to write event").Error())
		return
	}
}
