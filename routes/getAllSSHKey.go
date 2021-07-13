package routes

import (
	"net/http"

	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/utils"
)

func GetAllSSHKey(w http.ResponseWriter, r *http.Request) {
	sshKeyList, err := models.GetAllSSHKeyFromDB()
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
	}
	utils.JSON(w, http.StatusOK, sshKeyList)
}
