package routes

import (
	"net/http"

	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/utils"
)

// Get SSh connection from DB
func GetAllSSHConnection(w http.ResponseWriter, r *http.Request) {

	var sshConnection models.SshConnectionInfo
	sshConnectionList, err := sshConnection.GetAllSSHConnection()
	if err != nil {
		utils.JSON(w, http.StatusBadRequest, sshConnectionList)
		return
	}

	utils.JSON(w, http.StatusOK, sshConnectionList)

}
