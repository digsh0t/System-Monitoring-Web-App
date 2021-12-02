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

func GetWindowsLocalUserGroup(w http.ResponseWriter, r *http.Request) {

	type returnStruct struct {
		SshConnectionId int                     `json:"ssh_connection_id"`
		GroupList       []models.LocalUserGroup `json:"group_list"`
	}
	type inputStruct struct {
		SshIdList []int `json:"ssh_id_list"`
	}
	var returnVal []returnStruct
	var inputVal inputStruct

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

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	err = json.Unmarshal(body, &inputVal)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	for _, id := range inputVal.SshIdList {
		sshConnection, err := models.GetSSHConnectionFromId(id)
		if err != nil {
			utils.ERROR(w, http.StatusBadRequest, err.Error())
			return
		}
		groupList, err := sshConnection.GetLocalUserGroup()
		if err != nil {
			utils.ERROR(w, http.StatusBadRequest, err.Error())
			return
		}
		returnVal = append(returnVal, returnStruct{SshConnectionId: id, GroupList: groupList})
	}

	utils.JSON(w, http.StatusOK, returnVal)
}

func AddNewWindowsGroup(w http.ResponseWriter, r *http.Request) {

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

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}

	type unmarshalledGroup struct {
		SSHConnectionId []int  `json:"ssh_connection_id"`
		Name            string `json:"group_name"`
		Description     string `json:"description"`
	}
	var uG unmarshalledGroup

	json.Unmarshal(body, &uG)
	var hosts []string
	for _, id := range uG.SSHConnectionId {
		sshConnection, err := models.GetSSHConnectionFromId(id)
		if err != nil {
			utils.ERROR(w, http.StatusBadRequest, err.Error())
			return
		}
		hosts = append(hosts, sshConnection.HostNameSSH)
	}
	output, err := models.AddNewWindowsGroup(hosts, uG.Name, uG.Description)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}

	// Process Ansible Output
	status, fatal, err := models.ProcessingAnsibleOutput(output)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	returnJson := simplejson.New()
	returnJson.Set("Status", status)
	returnJson.Set("Fatal", fatal)
	utils.JSON(w, http.StatusOK, returnJson)

	// Write Event Web
	hostname, err := models.ConvertListIdToHostnameVer2(uG.SSHConnectionId)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, "fail to get hostname")
		return
	}
	description := "1 windows group added to host  [" + hostname + "]"
	_, err = models.WriteWebEvent(r, "Windows", description)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to write event").Error())
		return
	}
}

func RemoveWindowsGroup(w http.ResponseWriter, r *http.Request) {

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

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}

	type deletedGroup struct {
		SSHConnectionId int    `json:"ssh_connection_id"`
		Name            string `json:"group_name"`
	}

	var dG deletedGroup
	json.Unmarshal(body, &dG)

	sshConnection, err := models.GetSSHConnectionFromId(dG.SSHConnectionId)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}

	output, err := models.RemoveWindowsGroup(sshConnection.HostNameSSH, dG.Name)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	// Process Ansible Output
	status, fatal, err := models.ProcessingAnsibleOutput(output)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	returnJson := simplejson.New()
	returnJson.Set("Status", status)
	returnJson.Set("Fatal", fatal)
	utils.JSON(w, http.StatusOK, returnJson)

	// Write Event Web
	description := "1 windows group deleted from host  [" + sshConnection.HostNameSSH + "]"
	_, err = models.WriteWebEvent(r, "Windows", description)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to write event").Error())
		return
	}

}
