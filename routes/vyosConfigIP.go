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

func ConfigIPVyos(w http.ResponseWriter, r *http.Request) {

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

	var vyosJson models.VyOsJson
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to parse json").Error())
		return
	}
	err = json.Unmarshal(reqBody, &vyosJson)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to parse json").Error())
		return
	}

	// Config IP
	output, err := models.ConfigIPVyos(vyosJson)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get fatalList and recapList
	fatalList, recapList := models.RetrieveFatalRecap(output)

	// Parse recapList for analyzing
	recapInfoList, err := models.ParseRecap(recapList)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to parse recap").Error())
		return
	}

	// Analyzing recap
	result := models.AnalyzeRecap(recapInfoList)

	// Return Json
	returnJson := simplejson.New()
	returnJson.Set("FatalList", fatalList)
	returnJson.Set("Status", result)
	utils.JSON(w, http.StatusOK, returnJson)

}
