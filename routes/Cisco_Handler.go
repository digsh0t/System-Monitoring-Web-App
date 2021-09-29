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

func ListAllCisco(w http.ResponseWriter, r *http.Request) {
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

	sshConnectionList, err := models.ListAllCisco()
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to get list connection").Error())
	} else {
		utils.JSON(w, http.StatusOK, sshConnectionList)
	}

}

func GetInfoConfigCisco(w http.ResponseWriter, r *http.Request) {

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

	vars := mux.Vars(r)
	sshConnectionId, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to retrieve id").Error())
		return
	}
	configCiscoList, err := models.GetInfoConfigCisco(sshConnectionId)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
	} else {
		utils.JSON(w, http.StatusOK, configCiscoList)
	}

}

func GetInfoInterfaceCisco(w http.ResponseWriter, r *http.Request) {

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

	vars := mux.Vars(r)
	sshConnectionId, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to retrieve id").Error())
		return
	}
	interfaceCiscoList, err := models.GetInfoInterfaceCisco(sshConnectionId)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
	} else {
		utils.JSON(w, http.StatusOK, interfaceCiscoList)
	}

}

// config ipv4 and ipv6 for cisco
func ConfigIPCisco(w http.ResponseWriter, r *http.Request) {

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

	var ciscoJson models.CiscoJson
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to parse json").Error())
		return
	}
	err = json.Unmarshal(reqBody, &ciscoJson)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to parse json").Error())
		return
	}

	// Config IP
	output, err := models.ConfigIPCisco(ciscoJson)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}

	// Processing Output From Ansible
	status, fatals, err := models.ProcessingAnsibleOutput(output)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, "fail to process ansible output")
		return
	}

	// Return Json
	returnJson := simplejson.New()
	returnJson.Set("Status", status)
	returnJson.Set("Fatal", fatals)
	utils.JSON(w, http.StatusOK, returnJson)

	// Write Event Web
	hostname, err := models.ConvertListIdToHostnameVer2(ciscoJson.SshConnectionId)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, "fail to get hostname")
		return
	}
	description := "Config IP to network device " + hostname + " successfully"
	_, err = models.WriteWebEvent(r, "Network", description)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to write event").Error())
		return
	}

}

// config static route for cisco
func ConfigStaticRouteCisco(w http.ResponseWriter, r *http.Request) {

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

	var ciscoJson models.CiscoJson
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to parse json").Error())
		return
	}
	err = json.Unmarshal(reqBody, &ciscoJson)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to parse json").Error())
		return
	}

	// Config Static route
	output, err := models.ConfigStaticRouteCisco(ciscoJson)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}

	// Processing Output From Ansible
	status, fatals, err := models.ProcessingAnsibleOutput(output)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, "fail to process ansible output")
		return
	}

	// Return Json
	returnJson := simplejson.New()
	returnJson.Set("Status", status)
	returnJson.Set("Fatal", fatals)
	utils.JSON(w, http.StatusOK, returnJson)

	// Write Event Web
	hostname, err := models.ConvertListIdToHostnameVer2(ciscoJson.SshConnectionId)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, "fail to get hostname")
		return
	}
	description := "Config Static Route to network device " + hostname + " successfully"
	_, err = models.WriteWebEvent(r, "Network", description)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to write event").Error())
		return
	}

}

// Test Ping
func TestPingCisco(w http.ResponseWriter, r *http.Request) {

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

	var ciscoJson models.CiscoJson
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to parse json").Error())
		return
	}
	err = json.Unmarshal(reqBody, &ciscoJson)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to parse json").Error())
		return
	}

	// Test Ping
	output, err := models.TestPingCisco(ciscoJson)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}

	// Processing Output From Ansible
	status, fatals, err := models.ProcessingAnsibleOutput(output)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, "fail to process ansible output")
		return
	}

	// Return Json
	returnJson := simplejson.New()
	returnJson.Set("Status", status)
	returnJson.Set("Fatal", fatals)
	utils.JSON(w, http.StatusOK, returnJson)

	// Write Event Web
	hostname, err := models.ConvertListIdToHostnameVer2(ciscoJson.SshConnectionId)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, "fail to get hostname")
		return
	}
	description := "Test Ping to network device " + hostname + " successfully"
	_, err = models.WriteWebEvent(r, "Network", description)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to write event").Error())
		return
	}

}
