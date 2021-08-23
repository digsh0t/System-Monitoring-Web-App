package routes

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/bitly/go-simplejson"
	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/utils"
)

func LoadFile(w http.ResponseWriter, r *http.Request) {
	var (
		yaml models.YamlInfo
		err  error
	)

	reqBody, err := ioutil.ReadAll(r.Body)
	json.Unmarshal(reqBody, &yaml)

	// Call function Load in yaml.go
	_, err, fatalList, recapList := yaml.Load()

	// Return Json
	returnJson := simplejson.New()
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
