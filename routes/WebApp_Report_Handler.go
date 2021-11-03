package routes

import (
	"errors"
	"net/http"
	"time"

	"github.com/wintltr/login-api/auth"
	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/utils"
)

func GetReport(w http.ResponseWriter, r *http.Request, start time.Time) {

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

	report, err := models.GetReport(r, start)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
	} else {
		utils.JSON(w, http.StatusOK, report)
	}

}

func GetDetailOSReport(w http.ResponseWriter, r *http.Request) {

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
	ostype := query.Get("ostype")

	report, err := models.GetDetailOSReport(ostype)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
	} else {
		utils.JSON(w, http.StatusOK, report)
	}

}

func ExportReport(w http.ResponseWriter, r *http.Request) {

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
	filename := query.Get("filename")

	err = models.ExportReport(filename)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
	} else {
		utils.JSON(w, http.StatusOK, err)
	}

}
