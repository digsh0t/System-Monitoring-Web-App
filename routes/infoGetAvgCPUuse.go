package routes

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/bitly/go-simplejson"
	"github.com/gorilla/mux"
	"github.com/wintltr/login-api/auth"
	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/utils"
)

func GetAvgCpuUseRoute(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	if r.Method == "OPTIONS" {
		//CORS
		// return "OKOK"
		json.NewEncoder(w).Encode("OKOK")
		return
	}

	isAuthorized, err := auth.CheckAuth(r, []string{"admin", "user"})
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
	sshConnectionId, err := strconv.Atoi(vars["id"])
	if err != nil {
		returnJson.Set("Status", false)
		returnJson.Set("Error", errors.New("invalid SSH Connection id").Error())
		utils.JSON(w, http.StatusBadRequest, returnJson)
		return
	}
	sshConnection, err := models.GetSSHConnectionFromId(sshConnectionId)
	if err != nil {
		returnJson.Set("Status", false)
		returnJson.Set("Error", errors.New("fail to retrieve ssh connection").Error())
		utils.JSON(w, http.StatusBadRequest, returnJson)
		return
	}
	//dummy data
	command := "top -b -n 1"
	result, err := models.RunCommandFromSSHConnection(*sshConnection, command)
	if err != nil {
		returnJson.Set("Status", false)
		returnJson.Set("Error", errors.New("fail to get cpu usage").Error())
		utils.JSON(w, http.StatusBadRequest, returnJson)
		return
	}
	avgCPU, err := models.CalcAvgCPUFromTop(result)
	if err != nil {
		returnJson.Set("Status", false)
		returnJson.Set("Error", errors.New("fail to calculate cpu usage").Error())
		utils.JSON(w, http.StatusBadRequest, returnJson)
		return
	}
	utils.JSON(w, http.StatusOK, avgCPU)
}
