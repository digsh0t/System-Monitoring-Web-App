package routes

import (
	"errors"
	"net/http"

	"github.com/wintltr/login-api/auth"
	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/utils"
)

func GetReport(w http.ResponseWriter, r *http.Request) {

	//Authorization
	isAuthorized, err := auth.CheckAuth(r, []string{"admin"})
	if err != nil {
		utils.ERROR(w, http.StatusUnauthorized, errors.New("invalid token").Error())
		return
	}
	if !isAuthorized {
		utils.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized").Error())
		return
	}

	report, err := models.GetReport(r)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
	} else {
		utils.JSON(w, http.StatusOK, report)
	}

}
