package routes

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/utils"
)

func GetOSVersion(w http.ResponseWriter, r *http.Request) {
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
	osVersion, err := sshConnection.GetOSVersion()
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	utils.JSON(w, http.StatusOK, osVersion)
}

func GetCPUInfo(w http.ResponseWriter, r *http.Request) {
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
	cpu, err := sshConnection.GetCPUInfo()
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	utils.JSON(w, http.StatusOK, cpu)
}

func GetInterfaceList(w http.ResponseWriter, r *http.Request) {
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
	intefaceList, err := sshConnection.GetIntefaceList()
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	utils.JSON(w, http.StatusOK, intefaceList)
}
