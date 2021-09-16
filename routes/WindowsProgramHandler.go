package routes

import (
	"errors"
	"net/http"
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
