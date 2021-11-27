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

func LinuxClientGroupRemove(w http.ResponseWriter, r *http.Request) {

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
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to retrieve json format").Error())
		return
	}
	var groupJson models.LinuxClientGroupJson
	err = json.Unmarshal(reqBody, &groupJson)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, "fail to process json")
		return
	}

	var eventStatus string
	// Remove Host User
	output, err := models.LinuxClientGroupRemove(groupJson)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		eventStatus = "failed"
	} else {

		// Processing Output From Ansible
		status, fatalList, err := models.ProcessingAnsibleOutput(output)
		if err != nil {
			utils.ERROR(w, http.StatusBadRequest, "fail to process ansible output")
			return
		}

		// Return Json
		var statusCode int
		returnJson := simplejson.New()
		returnJson.Set("Status", status)
		returnJson.Set("Fatal", fatalList)
		if len(fatalList) > 0 {
			statusCode = http.StatusBadRequest
		} else {
			statusCode = http.StatusOK
		}
		utils.JSON(w, statusCode, returnJson)
		eventStatus = "successfully"
	}

	// Write Event Web
	hostnames, err := models.ConvertListIdToHostnameVer2(groupJson.SshConnectionIdList)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, "fail to conver list id to hostname")
		return
	}
	description := "User \"" + groupJson.Groupname + "\" removed from [" + hostnames + "] " + eventStatus
	_, err = models.WriteWebEvent(r, "LinuxGroup", description)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to write event").Error())
		return
	}
}

func LinuxClientGroupAdd(w http.ResponseWriter, r *http.Request) {

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
	if err != nil {
		utils.ERROR(w, http.StatusUnauthorized, errors.New("fail to parse json").Error())
		return
	}

	var groupJson models.LinuxClientGroupJson
	json.Unmarshal(reqBody, &groupJson)

	var eventStatus string
	output, err := models.LinuxClientGroupAdd(groupJson)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		eventStatus = "failed"
	} else {

		// Processing Output From Ansible
		status, fatalList, err := models.ProcessingAnsibleOutput(output)
		if err != nil {
			utils.ERROR(w, http.StatusBadRequest, "fail to process ansible output")
			return
		}

		// Return Json
		var statusCode int
		returnJson := simplejson.New()
		returnJson.Set("Status", status)
		returnJson.Set("Fatal", fatalList)
		if len(fatalList) > 0 {
			statusCode = http.StatusBadRequest
		} else {
			statusCode = http.StatusOK
		}
		utils.JSON(w, statusCode, returnJson)
		eventStatus = "successfully"
	}
	// Write Event Web
	hostnames, err := models.ConvertListIdToHostnameVer2(groupJson.SshConnectionIdList)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, "fail to conver list id to hostname")
		return
	}
	description := "Group \"" + groupJson.Groupname + "\" added to [" + hostnames + "] " + eventStatus
	_, err = models.WriteWebEvent(r, "LinuxGroup", description)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to write event").Error())
		return
	}
}

func LinuxClientGroupListAll(w http.ResponseWriter, r *http.Request) {

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

	// Retrieve Json Format
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.ERROR(w, http.StatusUnauthorized, errors.New("fail to parse json").Error())
		return
	}

	var groupJson models.LinuxClientGroupJson
	json.Unmarshal(reqBody, &groupJson)

	clientGroupList, err := models.LinuxClientGroupListAll(groupJson.SshConnectionIdList)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}

	// Return Json
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to get user from host").Error())
	} else {
		utils.JSON(w, http.StatusOK, clientGroupList)
	}

}
