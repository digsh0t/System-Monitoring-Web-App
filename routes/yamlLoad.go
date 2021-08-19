package routes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"

	"github.com/bitly/go-simplejson"
	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/utils"
)

func LoadFile(w http.ResponseWriter, r *http.Request) {

	returnJson := simplejson.New()
	var yaml models.YamlInfo

	reqBody, err := ioutil.ReadAll(r.Body)
	json.Unmarshal(reqBody, &yaml)

	// Establish command for load package
	command := "ansible-playbook ./yamls/" + yaml.File + " -e \"host=" + yaml.Host
	if yaml.Mode == "1" {
		command += " package=" + yaml.Package
	} else if yaml.Mode == "2" {
		command += " link=" + yaml.Link
	}
	command += "\""

	cmd := exec.Command("/bin/bash", "-c", command)
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	fmt.Println(out.String())
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
