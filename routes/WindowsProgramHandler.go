package routes

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/utils"
)

func GetWindowsInstalledProgram(w http.ResponseWriter, r *http.Request) {
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
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to read install info").Error())
		return
	}
	err = models.InstallWindowsProgram(string(body))
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to install program to client machine").Error())
		return
	}
}

func RemoveWindowsProgram(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to read program info").Error())
		return
	}

	var bodyMap map[string]interface{}
	json.Unmarshal(body, &bodyMap)
	regex, _ := regexp.Compile(`\{.*?\}`)
	programId := regex.FindString(fmt.Sprintf("%v", bodyMap["uninstall_string"]))
	err = models.DeleteWindowsProgram(bodyMap["host"], programId)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to remove program from client machine").Error())
		return
	}

}
