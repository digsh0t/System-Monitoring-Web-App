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
	returnJson := simplejson.New()
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("Fail to retrieve json format").Error())
		return
	}
	var hostUser models.HostUserInfo
	json.Unmarshal(reqBody, &hostUser)

	// Remove Host User
	ouput, err := hostUser.HostUserRemove()

	// Processing output and return Json
	var ansible models.AnsibleInfo
	fatalList, recapList := ansible.RetrieveFatalRecap(ouput)

	// Return Json
	var eventStatus string
	returnJson.Set("Fatal", fatalList)
	returnJson.Set("Recap", recapList)
	if err != nil {
		returnJson.Set("Status", false)
		returnJson.Set("Error", "Some client fail")
		eventStatus = "failed"
	} else {
		returnJson.Set("Status", true)
		returnJson.Set("Error", nil)
		eventStatus = "successfully"
	}
	utils.JSON(w, http.StatusBadRequest, returnJson)

	// Write Event Web
	host, err := ansible.ConvertListIdToHostname(hostUser.SshConnectionIdList)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("Fail to convert id string to hostname").Error())
		return
	}
	var description string
	if eventStatus == "failed" {
		description = "User \"" + hostUser.HostUserName + "\" removed from some host in list " + host + " " + eventStatus
	} else {
		description = "User \"" + hostUser.HostUserName + "\" removed from " + host + " " + eventStatus
	}
	_, err = models.WriteWebEvent(r, "HostUser", description)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("Fail to write event").Error())
		return
	}
}
