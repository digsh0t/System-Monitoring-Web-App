package routes

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/bitly/go-simplejson"
	"github.com/wintltr/login-api/auth"
	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/utils"
)

// create vlan switch
func CreateVlanSwitch(w http.ResponseWriter, r *http.Request) {

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

	var switchJson models.SwitchJson
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to parse json").Error())
		return
	}
	err = json.Unmarshal(reqBody, &switchJson)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to parse json").Error())
		return
	}

	// Config IP
	outputList, err := models.CreateVlanSwitch(switchJson)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}

	// Processing Output From Ansible
	status, fatals, err := models.ProcessingAnsibleOutputList(outputList)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, "fail to process ansible output")
		return
	}

	// Return Json
	returnJson := simplejson.New()
	var statusCode int
	if len(fatals) == 0 {
		statusCode = http.StatusOK
	} else {
		statusCode = http.StatusBadRequest
	}
	returnJson.Set("Status", status)
	returnJson.Set("Fatal", fatals)
	utils.JSON(w, statusCode, returnJson)

	// Write Event Web
	hostname, err := models.ConvertListIdToHostnameVer2(switchJson.SshConnectionId)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, "fail to get hostname")
		return
	}
	description := "Create vlan to network device " + hostname + " successfully"
	_, err = models.WriteWebEvent(r, "Network", description)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to write event").Error())
		return
	}

}

// get vlan switch
func GetVlanSwitch(w http.ResponseWriter, r *http.Request) {

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

	// Get Id parameter
	query := r.URL.Query()
	id, err := strconv.Atoi(query.Get("id"))
	if err != nil {
		utils.ERROR(w, http.StatusUnauthorized, errors.New("fail to convert id").Error())
		return
	}

	logs, err := models.GetVlanSwitch(id)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
	} else {
		utils.JSON(w, http.StatusOK, logs)
	}

}

// add interface to vlan
func AddInterfaceToVlanSwitch(w http.ResponseWriter, r *http.Request) {

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

	var switchJson models.SwitchJson
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to parse json").Error())
		return
	}
	err = json.Unmarshal(reqBody, &switchJson)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to parse json").Error())
		return
	}

	// add interfaces to vlan
	outputList, err := models.AddInterfacesToVlanSwitch(switchJson)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}

	// Processing Output From Ansible
	status, fatals, err := models.ProcessingAnsibleOutputList(outputList)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, "fail to process ansible output")
		return
	}

	// Return Json
	returnJson := simplejson.New()
	var statusCode int
	if len(fatals) == 0 {
		statusCode = http.StatusOK
	} else {
		statusCode = http.StatusBadRequest
	}
	returnJson.Set("Status", status)
	returnJson.Set("Fatal", fatals)
	utils.JSON(w, statusCode, returnJson)

	// Write Event Web
	hostname, err := models.ConvertListIdToHostnameVer2(switchJson.SshConnectionId)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, "fail to get hostname")
		return
	}
	description := "Add interfaces to vlan on network device " + hostname + " successfully"
	_, err = models.WriteWebEvent(r, "Network", description)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to write event").Error())
		return
	}

}

// delete vlan switch
func DeleteVlanSwitch(w http.ResponseWriter, r *http.Request) {

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

	var switchJson models.SwitchJson
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to parse json").Error())
		return
	}
	err = json.Unmarshal(reqBody, &switchJson)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to parse json").Error())
		return
	}

	// Config IP
	outputList, err := models.DeleteVlanSwitch(switchJson)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}

	// Processing Output From Ansible
	status, fatals, err := models.ProcessingAnsibleOutputList(outputList)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, "fail to process ansible output")
		return
	}

	// Return Json
	returnJson := simplejson.New()
	var statusCode int
	if len(fatals) == 0 {
		statusCode = http.StatusOK
	} else {
		statusCode = http.StatusBadRequest
	}
	returnJson.Set("Status", status)
	returnJson.Set("Fatal", fatals)
	utils.JSON(w, statusCode, returnJson)

	// Write Event Web
	hostname, err := models.ConvertListIdToHostnameVer2(switchJson.SshConnectionId)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, "fail to get hostname")
		return
	}
	description := "Delete vlan " + strconv.Itoa(switchJson.VlanId) + "from network device " + hostname + " successfully"
	_, err = models.WriteWebEvent(r, "Network", description)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to write event").Error())
		return
	}

}

// get interface switch
func GetInterfaceSwitch(w http.ResponseWriter, r *http.Request) {

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

	// Get Id parameter
	query := r.URL.Query()
	id, err := strconv.Atoi(query.Get("id"))
	if err != nil {
		utils.ERROR(w, http.StatusUnauthorized, errors.New("fail to convert id").Error())
		return
	}

	interfacesList, err := models.GetInterfaceSwitch(id)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
	} else {
		utils.JSON(w, http.StatusOK, interfacesList)
	}

}
