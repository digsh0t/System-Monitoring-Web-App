package routes

import (
	"net/http"

	"github.com/bitly/go-simplejson"
	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/utils"
)

func TestSSHConnection(w http.ResponseWriter, r *http.Request) {
	var sshConnection models.SshConnectionInfo

	//Dummy Database SSH Connection Info
	sshConnection.HostSSH = "192.168.163.136"
	sshConnection.UserSSH = "root"
	sshConnection.PortSSH = 22

	success, err := sshConnection.TestConnectionPublicKey()

	returnJson := simplejson.New()
	returnJson.Set("Status", success)
	returnJson.Set("Error", err.Error())
	if err != nil {
		utils.JSON(w, http.StatusBadRequest, returnJson)
	} else {
		utils.JSON(w, http.StatusOK, returnJson)
	}
}
