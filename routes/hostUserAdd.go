package routes

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/bitly/go-simplejson"
	"github.com/wintltr/login-api/auth"
	"github.com/wintltr/login-api/event"
	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/utils"
)

func HostUserAdd(w http.ResponseWriter, r *http.Request) {

	// Authorization
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
	returnJson := simplejson.New()
	if err != nil {
		returnJson.Set("Status", false)
		returnJson.Set("Error", "Fail to retrieve Json format")
		utils.JSON(w, http.StatusBadRequest, returnJson)
		return
	}
	var hostUser models.HostUserInfo
	json.Unmarshal(reqBody, &hostUser)

	// Add Host User
	ouput, err := hostUser.HostUserAdd()

	// Write Event Web
	var ansible models.AnsibleInfo
	host, err := ansible.ConvertListIdToHostname(hostUser.SshConnectionId)
	if err != nil {
		returnJson.Set("Status", false)
		returnJson.Set("Error", "Fail to convert id to hostname")
		utils.JSON(w, http.StatusBadRequest, returnJson)
		return
	}

	description := "Add host user to " + host
	_, err = event.WriteWebEvent(r, "HostUser", description)
	if err != nil {
		returnJson.Set("Status", false)
		returnJson.Set("Error", "Fail to write web event")
		utils.JSON(w, http.StatusBadRequest, returnJson)
		return
	}

	// Processing output and return Json
	fatalList, recapList := ansible.RetrieveFatalRecap(ouput)

	returnJson.Set("Fatal", fatalList)
	returnJson.Set("Recap", recapList)
	if err != nil {
		returnJson.Set("Status", false)
		returnJson.Set("Error", "Some client fail")
	} else {
		returnJson.Set("Status", true)
		returnJson.Set("Error", nil)
	}
	utils.JSON(w, http.StatusBadRequest, returnJson)
	return
}
