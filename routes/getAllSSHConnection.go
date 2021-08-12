package routes

import (
	"net/http"

	"github.com/wintltr/login-api/auth"
	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/utils"
)

// Get SSh connection from DB
func GetAllSSHConnection(w http.ResponseWriter, r *http.Request) {

	isAuthorized, err := auth.CheckAuth(r, []string{"admin", "user"})
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	if !isAuthorized {
		utils.ERROR(w, http.StatusUnauthorized, err.Error())
		return
	}

	var sshConnection models.SshConnectionInfo
	sshConnectionList, err := sshConnection.GetAllSSHConnection()
	if err != nil {
		utils.JSON(w, http.StatusBadRequest, sshConnectionList)
		return
	}

	utils.JSON(w, http.StatusOK, sshConnectionList)

}
