package routes

import (
	"net/http"

	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/utils"
)

func GetAllDefaultIP(w http.ResponseWriter, r *http.Request) {

	defaultIPList, err := models.GetAllDefaultIP()
	if err != nil {
		utils.JSON(w, http.StatusBadRequest, defaultIPList)
		return
	}
	utils.JSON(w, http.StatusOK, defaultIPList)

}
