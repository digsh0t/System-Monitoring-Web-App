package routes

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/utils"
)

func GetWindowsFirewall(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	direction := vars["direction"]
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to read ssh connection id").Error())
		return
	}
	sshConnection, err := models.GetSSHConnectionFromId(id)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to prepare ssh connection").Error())
		return
	}
	firewallRules, err := sshConnection.GetWindowsFirewall(direction)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to read windows firewall rules from ssh connection").Error())
		return
	}
	utils.JSON(w, http.StatusOK, firewallRules)
}
