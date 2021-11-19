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

func GetWindowsLocalUserGroup(w http.ResponseWriter, r *http.Request) {

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
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
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
	utils.JSON(w, http.StatusOK, groupList)
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

}
