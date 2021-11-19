package routes

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/wintltr/login-api/auth"
	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/utils"
)

func GetWindowsEventLogs(w http.ResponseWriter, r *http.Request) {

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

	// Get Id parameter
	query := r.URL.Query()
	id, err := strconv.Atoi(query.Get("id"))
	if err != nil {
		utils.ERROR(w, http.StatusUnauthorized, errors.New("fail to convert id").Error())
		return
	}
	// Get logname parameter
	logname := query.Get("logname")

	// Get Start Time parameter
	startTime := query.Get("start")

	// Get End Time parameter
	endTime := query.Get("end")

	windowsLogsList, err := models.GetWindowsEventLogs(id, logname, startTime, endTime)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
	} else {
		utils.JSON(w, http.StatusOK, windowsLogsList)
	}

}

func GetDetailWindowsEventLog(w http.ResponseWriter, r *http.Request) {

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

	// Get Id parameter
	query := r.URL.Query()
	id, err := strconv.Atoi(query.Get("id"))
	if err != nil {
		utils.ERROR(w, http.StatusUnauthorized, errors.New("fail to convert id").Error())
		return
	}
	// Get logname parameter
	logname := query.Get("logname")

	// Get Index parameter
	index := query.Get("index")

	windowsLogsList, err := models.GetDetailWindowsEventLog(id, logname, index)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
	} else {
		utils.JSON(w, http.StatusOK, windowsLogsList)
	}

}
