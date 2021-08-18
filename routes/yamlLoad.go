package routes

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os/exec"

	"github.com/bitly/go-simplejson"
	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/utils"
)

func LoadFile(w http.ResponseWriter, r *http.Request) {

	returnJson := simplejson.New()
	var yamlInfo models.YamlInfo

	reqBody, err := ioutil.ReadAll(r.Body)
	json.Unmarshal(reqBody, &yamlInfo)
	cmd := exec.Command("ansible-playbook", "./yamls/"+yamlInfo.FileName, "-e", "host="+yamlInfo.Host)
	err = cmd.Run()
	if err != nil {
		if err.Error() == "exit status 1" {
			returnJson.Set("Status", false)
			returnJson.Set("Error", "File name not found!")
			utils.JSON(w, http.StatusBadRequest, returnJson)
			return
		}
	}

	returnJson.Set("Status", true)
	returnJson.Set("Error", nil)
	utils.JSON(w, http.StatusOK, returnJson)

}
