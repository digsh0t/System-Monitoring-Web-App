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
	var ansible models.AnsibleInfo
	hostStr, err := ansible.ProcessingHost(packages.Host)
	if err != nil {
		returnJson.Set("Status", false)
		returnJson.Set("Error", "Fail to processing list host")
		utils.JSON(w, http.StatusBadRequest, returnJson)
		return
	}

	if packages.Mode == "1" {
		// Install By Name
		ansible.ExtraValue = map[string]string{"host": hostStr, "package": packages.Package}
	} else if packages.Mode == "2" {
		// Install By Link
		ansible.ExtraValue = map[string]string{"host": hostStr, "link": packages.Link}
	}
	output, err := ansible.Load("./yamls/" + packages.File)

	// Processing Output From Ansible
	fatalList, recapList := ansible.ProcessingOutput(output)
	var recapStruct models.RecapInfo
	recapStructList, errRecap := recapStruct.ProcessingRecap(recapList)
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
	returnJson.Set("Fatal", fatalList)
	returnJson.Set("Recap", recapList)
	if err != nil {
		returnJson.Set("Status", false)
		returnJson.Set("Error", err.Error())

	} else {
		returnJson.Set("Status", true)
		returnJson.Set("Error", nil)
	}

	utils.JSON(w, http.StatusOK, returnJson)
	return

}
