package routes

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/bitly/go-simplejson"
	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/utils"
)

func PackageInstall(w http.ResponseWriter, r *http.Request) {

	/*
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			//CORS
			// return "OKOK"
			json.NewEncoder(w).Encode("OKOK")
			return
		}

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

	*/
	var yaml models.YamlInfo
	returnJson := simplejson.New()

	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		returnJson.Set("Status", false)
		returnJson.Set("Error", "Fail to retrieve json format")
		utils.JSON(w, http.StatusBadRequest, returnJson)
		return
	}
	json.Unmarshal(reqBody, &yaml)

	// Call function Load in yaml.go
	_, err, fatalList, recapList := yaml.Load()

	var recapStruct models.RecapInfo
	recapStructList, errRecap := recapStruct.ProcessingRecap(recapList)
	if errRecap != nil {
		utils.JSON(w, http.StatusBadRequest, errRecap.Error())
		return
	}
	for _, node := range recapStructList {
		fmt.Println(node)
	}

	// Return Json
	if err != nil {
		returnJson.Set("Status", true)
		returnJson.Set("Error", err.Error())
		returnJson.Set("Fatal", fatalList)
		returnJson.Set("Recap", recapList)
		utils.JSON(w, http.StatusOK, returnJson)
		return

	}

	returnJson.Set("Status", true)
	returnJson.Set("Error", nil)
	utils.JSON(w, http.StatusOK, returnJson)
	return

}
