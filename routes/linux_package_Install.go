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
	var packages models.PackageJson
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.ERROR(w, http.StatusUnauthorized, errors.New("fail to parse json").Error())
		return
	}
	json.Unmarshal(reqBody, &packages)

	// Load File Yaml Install
	hostStr, err := models.ConvertListIdToHostnameVer2(packages.Host)
	if err != nil {
		utils.ERROR(w, http.StatusUnauthorized, errors.New("fail to convert id").Error())
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
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, "fail to load yaml")
		return
	}

	// Processing Output From Ansible
	status, fatalList, err := models.ProcessingAnsibleOutput(output)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, "fail to process ansible output")
		return
	}

	// Return Json
	returnJson := simplejson.New()
	returnJson.Set("Fatal", fatalList)
	returnJson.Set("Status", status)
	utils.JSON(w, http.StatusOK, returnJson)

	// Write Event Web
	description := "Package \"" + packages.Package + "\" installed to " + hostStr + " successfully"
	_, err = models.WriteWebEvent(r, "Package", description)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to write event").Error())
		return
	}

}
