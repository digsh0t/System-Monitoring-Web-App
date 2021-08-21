package routes

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os/exec"
	"regexp"
	"strings"

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
	raw := out.String()

	// Return Json
	returnJson := simplejson.New()
	if err != nil {
		fatalList, recapList := ProcessingOutput(raw)
		returnJson.Set("Status", true)
		returnJson.Set("Error", err)
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

func ProcessingOutput(raw string) ([]string, []string) {
	var fatalList []string
	var recapList []string

	// Extracting Fatal
	text := strings.Split(raw, "\n")
	for _, line := range text {
		pattern := "^fatal"
		r, _ := regexp.Compile(pattern)
		if r.MatchString(line) {
			fatalList = append(fatalList, line)
		}
	}

	// Extracting PLAY RECAP **********
	pattern := "PLAY RECAP .+\n"
	r, _ := regexp.Compile(pattern)
	strIndex := r.FindStringIndex(raw)
	tmp := raw[strIndex[1]:]
	text = strings.Split(tmp, "\n")
	for _, line := range text {
		if line != "" {
			recapList = append(recapList, line)
		}
	}

	return fatalList, recapList

}
