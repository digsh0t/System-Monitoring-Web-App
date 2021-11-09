package routes

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

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

func GetAllClientSysLogRoute(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	date := vars["date"]
	logs, err := models.GetAllClientSyslog("/var/log/remotelogs", date)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	if logs == nil {
		utils.JSON(w, http.StatusOK, logs)
		return
	}
	utils.JSON(w, http.StatusOK, logs[:10])
}

func SetupSyslogRoute(w http.ResponseWriter, r *http.Request) {

	type receivedStruct struct {
		ID       int    `json:"id"`
		ServerIP string `json:"server_ip"`
	}
	var received receivedStruct
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	err = json.Unmarshal(body, &received)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}

	sshConnection, err := models.GetSSHConnectionFromId(received.ID)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	if strings.Contains(strings.ToLower(sshConnection.OsType), "windows") {
		err = sshConnection.InstallNxlogWindows()
		if err != nil {
			utils.ERROR(w, http.StatusBadRequest, err.Error())
			return
		}
		err = sshConnection.SetupSyslogWindows(received.ServerIP, `C:\Program Files (x86)\nxlog\conf\nxlog.conf`)
	} else if strings.Contains(strings.ToLower(sshConnection.OsType), "ubuntu") {
		err = sshConnection.InstallRsyslog()
		if err != nil {
			utils.ERROR(w, http.StatusBadRequest, err.Error())
			return
		}
		_, err = sshConnection.SetupSyslogRsyslog(received.ServerIP, `/etc/rsyslog.conf`)
	}
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
}
