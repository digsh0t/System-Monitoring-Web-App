package routes

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/bitly/go-simplejson"
	"github.com/gorilla/mux"
	"github.com/wintltr/login-api/auth"
	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/utils"
)

func SSHKeyDeleteRoute(w http.ResponseWriter, r *http.Request) {

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

	returnJson := simplejson.New()
	vars := mux.Vars(r)
	sshKeyId, err := strconv.Atoi(vars["id"])
	if err != nil {
		returnJson.Set("Status", false)
		returnJson.Set("Error", errors.New("invalid SSH key id").Error())
		utils.JSON(w, http.StatusBadRequest, returnJson)
		return
	}

	_, err = models.SSHKeyDelete(sshKeyId)
	if err != nil {
		returnJson.Set("Status", false)
		returnJson.Set("Error", errors.New("please delete ssh connections associated with this SSH key first").Error())
		utils.JSON(w, http.StatusBadRequest, returnJson)
		return
	}

	returnJson.Set("Status", true)
	returnJson.Set("Error", nil)
	utils.JSON(w, http.StatusOK, returnJson)

}