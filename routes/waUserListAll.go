package routes

import (
	"errors"
	"net/http"

	"github.com/wintltr/login-api/auth"
	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/utils"
)

func ListAllWebAppUser(w http.ResponseWriter, r *http.Request) {
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

	var waUserList []models.User
	waUserList, err = models.ListAllWepAppUser()

	// Return json
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("Fail to get all list users").Error())
	} else {
		utils.JSON(w, http.StatusOK, waUserList)
	}

}
