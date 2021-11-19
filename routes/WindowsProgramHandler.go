package routes

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"

	"github.com/bitly/go-simplejson"
	"github.com/gorilla/mux"
	"github.com/wintltr/login-api/auth"
	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/utils"
)

func GetWindowsInstalledProgram(w http.ResponseWriter, r *http.Request) {

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
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to get ssh connection id").Error())
		return
	}
	sshConnection, err := models.GetSSHConnectionFromId(sshConnectionId)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to get ssh connection from provided id").Error())
		return
	}
	installedPrograms, err := sshConnection.GetInstalledProgram()
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to get installed programs from client machine").Error())
		return
	}
	utils.JSON(w, http.StatusOK, installedPrograms)
}

func InstallWindowsProgram(w http.ResponseWriter, r *http.Request) {

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
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to read install info").Error())
		return
	}

	type unmarshalledProgram struct {
		SSHConnectionId []int  `json:"ssh_connection_id"`
		URL             string `json:"url"`
		Dest            string `json:"dest"`
	}
	var uP unmarshalledProgram

	json.Unmarshal(body, &uP)
	var hosts []string
	for _, id := range uP.SSHConnectionId {
		sshConnection, err := models.GetSSHConnectionFromId(id)
		if err != nil {
			utils.ERROR(w, http.StatusBadRequest, err.Error())
			return
		}
		hosts = append(hosts, sshConnection.HostNameSSH)
	}
	output, err := models.InstallWindowsProgram(hosts, uP.URL, uP.Dest)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to install program to client machine").Error())
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

func RemoveWindowsProgram(w http.ResponseWriter, r *http.Request) {

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
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to read program info").Error())
		return
	}

	type deletedProgram struct {
		SSHConnectionId int    `json:"ssh_connection_id"`
		UninstallString string `json:"uninstall_string"`
	}
	var dP deletedProgram
	json.Unmarshal(body, &dP)

	sshConnection, err := models.GetSSHConnectionFromId(dP.SSHConnectionId)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to get machine info from provided id").Error())
		return
	}

	regex, _ := regexp.Compile(`\{.*?\}`)
	programId := regex.FindString(dP.UninstallString)
	output, err := models.DeleteWindowsProgram(sshConnection.HostNameSSH, programId)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to remove program from client machine").Error())
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
