package routes

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/bitly/go-simplejson"
	"github.com/gorilla/mux"
	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/utils"
)

func GetSysLogFilesRoute(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	date := vars["date"]
	page, err := strconv.Atoi(vars["page"])
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	logs, err := models.GetClientSyslog("/var/log/remotelogs", id, date)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	totalPage := (len(logs) / 100) + 1
	returnJson := simplejson.New()
	returnJson.Set("pages", totalPage)
	if totalPage < page {
		utils.ERROR(w, http.StatusBadRequest, errors.New("invalid page").Error())
		return
	}
	if len(logs) < 100*page {
		logs = logs[100*(page-1):]
	} else {
		logs = logs[100*(page-1) : 100*page]
	}
	returnJson.Set("logs", logs)
	utils.JSON(w, http.StatusOK, returnJson)
}

func GetSysLogByPriRoute(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	pri, err := strconv.Atoi(vars["pri"])
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	date := vars["date"]
	page, err := strconv.Atoi(vars["page"])
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	logs, err := models.GetClientSyslogByPri("/var/log/remotelogs", id, date, pri)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	totalPage := (len(logs) / 100) + 1
	returnJson := simplejson.New()
	returnJson.Set("pages", totalPage)
	if totalPage < page {
		utils.ERROR(w, http.StatusBadRequest, errors.New("invalid page").Error())
		return
	}
	if len(logs) < 100*page {
		logs = logs[100*(page-1):]
	} else {
		logs = logs[100*(page-1) : 100*page]
	}
	returnJson.Set("logs", logs)
	utils.JSON(w, http.StatusOK, returnJson)
}

func GetAllClientSysLogPriStatRoute(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	date := vars["date"]
	logs, err := models.GetAllClientSyslogPriStat("/var/log/remotelogs", date)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	utils.JSON(w, http.StatusOK, logs)
}
