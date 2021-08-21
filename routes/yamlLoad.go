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
	var (
		hostStr string
		yaml    models.YamlInfo
		out     bytes.Buffer
	)

	reqBody, err := ioutil.ReadAll(r.Body)
	json.Unmarshal(reqBody, &yaml)

	// Processing a list host
	for _, v := range yaml.Host {
		hostStr += v + ","
	}

	// Establish command for load package
	command := "ansible-playbook ./yamls/" + yaml.File + " -e \"host=" + hostStr
	if yaml.Mode == "1" {
		command += " package=" + yaml.Package
	} else if yaml.Mode == "2" {
		command += " link=" + yaml.Link
	}
	command += "\""

	cmd := exec.Command("/bin/bash", "-c", command)
	cmd.Stdout = &out
	err = cmd.Run()
	fmt.Println(out.String())

	// Return Json
	returnJson := simplejson.New()
	if err != nil {
		fmt.Println("error")
	}

	returnJson.Set("Status", true)
	returnJson.Set("Error", nil)
	utils.JSON(w, http.StatusOK, returnJson)

}
