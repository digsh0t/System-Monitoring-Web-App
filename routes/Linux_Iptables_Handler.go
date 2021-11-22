package routes

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/bitly/go-simplejson"
	"github.com/gorilla/mux"
	"github.com/wintltr/login-api/auth"
	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/utils"
)

func LinuxClientIptablesRemove(w http.ResponseWriter, r *http.Request) {

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
	var iptablesJson models.IptablesJson
	err = json.Unmarshal(reqBody, &iptablesJson)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, "fail to process json")
		return
	}

	// Remove Host User
	var eventStatus string
	output, err := models.LinuxClientIptablesRemove(iptablesJson)
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
		returnJson := simplejson.New()
		returnJson.Set("Status", status)
		returnJson.Set("Fatal", fatalList)
		utils.JSON(w, http.StatusOK, returnJson)
		eventStatus = "successfully"
	}

	// Write Event Web
	hostnames, err := models.ConvertListIdToHostnameVer2(iptablesJson.SshConnectionId)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, "fail to conver list id to hostname")
		return
	}
	description := "1 Rule removed successfully from machines [" + hostnames + "] " + eventStatus
	_, err = models.WriteWebEvent(r, "LinuxUser", description)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to write event").Error())
		return
	}
}

func LinuxClientIptablesAdd(w http.ResponseWriter, r *http.Request) {

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

	var iptablesJson models.IptablesJson
	json.Unmarshal(reqBody, &iptablesJson)

	var eventStatus string
	output, err := models.LinuxClientIptablesAdd(iptablesJson)
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
		returnJson := simplejson.New()
		returnJson.Set("Status", status)
		returnJson.Set("Fatal", fatalList)
		utils.JSON(w, http.StatusOK, returnJson)
		eventStatus = "successfully"
	}
	// Write Event Web
	hostnames, err := models.ConvertListIdToHostnameVer2(iptablesJson.SshConnectionId)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, "fail to conver list id to hostname")
		return
	}
	description := "1 Rule added successfully to machines [" + hostnames + "] " + eventStatus
	_, err = models.WriteWebEvent(r, "LinuxIptables", description)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to write event").Error())
		return
	}
}

func LinuxClientIptablesListAll(w http.ResponseWriter, r *http.Request) {

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

	vars := mux.Vars(r)
	stringId := vars["id"]
	intId, err := strconv.Atoi(stringId)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to convert id string to int").Error())
		return
	}
	clientIptablesList, err := models.LinuxClientIptablesListAll(intId)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}

	// Return Json
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to get iptables from host").Error())
	} else {
		utils.JSON(w, http.StatusOK, clientIptablesList)
	}

}
