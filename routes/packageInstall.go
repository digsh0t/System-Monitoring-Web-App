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

func PackageInstall(w http.ResponseWriter, r *http.Request) {

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

	// Retrieve Json Format
	var packages models.PackageInfo
	reqBody, err := ioutil.ReadAll(r.Body)
	returnJson := simplejson.New()
	if err != nil {
		returnJson.Set("Status", false)
		returnJson.Set("Error", "Fail to retrieve json format")
		utils.JSON(w, http.StatusBadRequest, returnJson)
		return
	}
	json.Unmarshal(reqBody, &packages)

	// Load File Yaml Install
	hostStr, err := models.ConvertListIdToHostname(packages.Host)
	if err != nil {
		returnJson.Set("Status", false)
		returnJson.Set("Error", "Fail to processing list host")
		utils.JSON(w, http.StatusBadRequest, returnJson)
		return
	}

	var extraValue map[string]string
	if packages.Mode == "1" {
		// Install By Name
		extraValue = map[string]string{"host": hostStr, "package": packages.Package}
	} else if packages.Mode == "2" {
		// Install By Link
		extraValue = map[string]string{"host": hostStr, "link": packages.Link}
	}
	output, err := models.LoadYAML("./yamls/"+packages.File, extraValue)

	// Processing Output From Ansible
	fatalList, recapList := models.RetrieveFatalRecap(output)
	recapStructList, errRecap := models.ParseRecap(recapList)
	if errRecap != nil {
		returnJson.Set("Status", false)
		returnJson.Set("Error", "Fail to process output from ansible")
		utils.JSON(w, http.StatusBadRequest, returnJson)
		return
	}

	// Add Successful Installed Package to DB
	if packages.Mode == "1" {
		_, errPackge := models.AddPackage(recapStructList, packages.Package)
		if errPackge != nil {
			returnJson.Set("Status", false)
			returnJson.Set("Error", "Fail to add installed package to db ")
			utils.JSON(w, http.StatusBadRequest, returnJson)
			return
		}
	}

	// Return Json
	var eventStatus string
	returnJson.Set("Fatal", fatalList)
	returnJson.Set("Recap", recapList)
	if err != nil {
		returnJson.Set("Status", false)
		returnJson.Set("Error", err.Error())
		eventStatus = "failed"
	} else {
		returnJson.Set("Status", true)
		returnJson.Set("Error", nil)
		eventStatus = "successfully"
	}
	utils.JSON(w, http.StatusOK, returnJson)

	// Write Event Web
	var description string
	if eventStatus == "failed" {
		description = "Package \"" + packages.Package + "\" installed to some host in list " + hostStr + " " + eventStatus
	} else {
		description = "Package \"" + packages.Package + "\" installed to " + hostStr + " " + eventStatus
	}
	_, err = models.WriteWebEvent(r, "Package", description)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("Fail to write event").Error())
		return
	}
	return

}
