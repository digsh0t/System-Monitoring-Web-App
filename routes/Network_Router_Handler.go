package routes

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/bitly/go-simplejson"
	"github.com/wintltr/login-api/auth"
	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/utils"
)

// config Ipv4 for router
func ConfigIPRouter(w http.ResponseWriter, r *http.Request) {

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

	var routerJson models.RouterJson
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to parse json").Error())
		return
	}
	err = json.Unmarshal(reqBody, &routerJson)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to parse json").Error())
		return
	}

	// Config IP
	var eventStatus string
	outputList, err := models.ConfigIPRouter(routerJson)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		eventStatus = "failed"
	} else {

		// Processing Output From Ansible
		status, fatals, err := models.ProcessingAnsibleOutputList(outputList)
		if err != nil {
			utils.ERROR(w, http.StatusBadRequest, "fail to process ansible output")
			return
		}

		// Return Json
		returnJson := simplejson.New()
		var statusCode int
		if len(fatals) > 0 {
			statusCode = http.StatusBadRequest
		} else {
			statusCode = http.StatusOK
		}
		returnJson.Set("Status", status)
		returnJson.Set("Fatal", fatals)
		utils.JSON(w, statusCode, returnJson)
		eventStatus = "successfully"
	}

	// Write Event Web
	hostname, err := models.ConvertListIdToHostnameVer2(routerJson.SshConnectionId)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, "fail to get hostname")
		return
	}
	description := "Config IP to network device [" + hostname + "] " + eventStatus
	_, err = models.WriteWebEvent(r, "Network", description)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to write event").Error())
		return
	}

}

// config static route for router
func ConfigStaticRouteRouter(w http.ResponseWriter, r *http.Request) {

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

	var routerJson models.RouterJson
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to parse json").Error())
		return
	}
	err = json.Unmarshal(reqBody, &routerJson)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to parse json").Error())
		return
	}

	// Config IP
	var eventStatus string
	outputList, err := models.ConfigStaticRouteRouter(routerJson)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		eventStatus = "failed"
	} else {
		// Processing Output From Ansible
		status, fatals, err := models.ProcessingAnsibleOutputList(outputList)
		if err != nil {
			utils.ERROR(w, http.StatusBadRequest, "fail to process ansible output")
			return
		}

		// Return Json
		var statusCode int
		returnJson := simplejson.New()
		returnJson.Set("Status", status)
		returnJson.Set("Fatal", fatals)
		if len(fatals) > 0 {
			statusCode = http.StatusBadRequest
		} else {
			statusCode = http.StatusOK
		}
		utils.JSON(w, statusCode, returnJson)
		eventStatus = "successfully"
	}

	// Write Event Web
	hostname, err := models.ConvertListIdToHostnameVer2(routerJson.SshConnectionId)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, "fail to get hostname")
		return
	}
	description := "Config static route to network device [" + hostname + "] " + eventStatus
	_, err = models.WriteWebEvent(r, "Network", description)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to write event").Error())
		return
	}

}
