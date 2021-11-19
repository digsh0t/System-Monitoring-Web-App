package routes

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/wintltr/login-api/auth"
	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/utils"
)

func HistoryListAll(w http.ResponseWriter, r *http.Request) {
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

	vars := mux.Vars(r)
	stringId := vars["id"]
	intId, err := strconv.Atoi(stringId)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("Fail to convert id string to int").Error())
		return
	}

	historyList, err := models.HistoryListAll(intId)

	// Return Json
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("Fail to user command history from host").Error())
	} else {
		utils.JSON(w, http.StatusOK, historyList)
	}

}
