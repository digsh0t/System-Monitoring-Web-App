package routes

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/bitly/go-simplejson"
	"github.com/gorilla/mux"
	"github.com/wintltr/login-api/auth"
	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/utils"
)

func SSHConnectionDeleteRoute(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	if r.Method == "OPTIONS" {
		//CORS
		// return "OKOK"
		json.NewEncoder(w).Encode("OKOK")
		return
	}

	isAuthorized, err := auth.CheckAuth(r, []string{"admin"})
	if err != nil {
		utils.ERROR(w, http.StatusUnauthorized, errors.New("invalid token").Error())
		return
	}
	if !isAuthorized {
		utils.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized").Error())
		return
	}

	returnJson := simplejson.New()
	vars := mux.Vars(r)
	sshConnectionId, err := strconv.Atoi(vars["id"])
	if err != nil {
		returnJson.Set("Status", false)
		returnJson.Set("Error", errors.New("invalid SSH Connection id").Error())
		utils.JSON(w, http.StatusBadRequest, returnJson)
		return
	}

	sshConnectionInfo, _ := models.GetSSHConnectionFromId(sshConnectionId)
	if sshConnectionInfo == nil {
		returnJson.Set("Status", false)
		returnJson.Set("Error", errors.New("SSH Connection with id "+strconv.Itoa(sshConnectionId)+" doesn't exist").Error())
		utils.JSON(w, http.StatusBadRequest, returnJson)
		return
	}
	cmd := `echo "y" | if grep -v key` + strconv.Itoa(sshConnectionInfo.SSHKeyId) + ` $HOME/.ssh/authorized_keys > $HOME/.ssh/tmp; then cat $HOME/.ssh/tmp > $HOME/.ssh/authorized_keys && rm $HOME/.ssh/tmp; fi;`
	_, err = models.ExecCommand(cmd, sshConnectionInfo.UserSSH, sshConnectionInfo.PasswordSSH, sshConnectionInfo.HostSSH, sshConnectionInfo.PortSSH)
	if err != nil {
		returnJson.Set("Status", false)
		returnJson.Set("Error", errors.New("error while removing SSH Connection key from remote server").Error())
		utils.JSON(w, http.StatusBadRequest, returnJson)
		return
	}

	_, err = models.DeleteSSHConnection(sshConnectionId)
	if err != nil {
		returnJson.Set("Status", false)
		returnJson.Set("Error", errors.New("error while deleting SSH Connection").Error())
		utils.JSON(w, http.StatusBadRequest, returnJson)
		return
	}
	err = models.GenerateInventory()
	if err != nil {
		returnJson.Set("Status", false)
		returnJson.Set("Error", errors.New("error while regenerate ansible inventory").Error())
		utils.JSON(w, http.StatusBadRequest, returnJson)
		return
	}

	event := models.Event{
		EventType:   "SSH Connection",
		Description: "SSH Connection to " + sshConnectionInfo.HostNameSSH + " deleted",
		TimeStampt:  time.Now(),
		CreatorId:   sshConnectionInfo.CreatorId,
	}
	models.CreateEvent(event)

	returnJson.Set("Status", true)
	returnJson.Set("Error", nil)
	utils.JSON(w, http.StatusOK, returnJson)
}
