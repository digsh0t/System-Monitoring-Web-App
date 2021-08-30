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

func PackageRemove(w http.ResponseWriter, r *http.Request) {

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
	returnJson := simplejson.New()
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		returnJson.Set("Status", false)
		returnJson.Set("Error", "Fail to retrieve json format")
		utils.JSON(w, http.StatusBadRequest, returnJson)
		return
	}
	json.Unmarshal(reqBody, &packages)

	// Call function Load in yaml.go
	var ansible models.AnsibleInfo
	hostStr, err := ansible.ConvertListIdToHostname(packages.Host)
	if err != nil {
		returnJson.Set("Status", false)
		returnJson.Set("Error", "Fail to processing list host")
		utils.JSON(w, http.StatusBadRequest, returnJson)
		return
	}
	ansible.ExtraValue = map[string]string{"host": hostStr, "package": packages.Package}
	output, err := ansible.Load("./yamls/" + packages.File)

	// Processing Output From Ansible
	fatalList, recapList := ansible.RetrieveFatalRecap(output)
	var recapStruct models.RecapInfo
	recapStructList, errRecap := recapStruct.ProcessingRecap(recapList)
	if errRecap != nil {
		returnJson.Set("Status", false)
		returnJson.Set("Error", "Fail to process output from ansible")
		utils.JSON(w, http.StatusBadRequest, returnJson)
		return
	}

	// Remove package from DB
	_, errPackge := models.RemovePackage(recapStructList, packages.Package)
	if errPackge != nil {
		returnJson.Set("Status", false)
		returnJson.Set("Error", "Fail to remove package to db ")
		utils.JSON(w, http.StatusBadRequest, returnJson)
		return
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
