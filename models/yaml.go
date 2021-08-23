package models

import (
	"bytes"
	"os/exec"
	"regexp"
	"strings"
)

type YamlInfo struct {
	Host    []string `json:"host"`
	File    string   `json:"file"`
	Mode    string   `json:"mode"`
	Package string   `json:"package"`
	Link    string   `json:"link"`
}

func (yaml *YamlInfo) Load() (string, error, []string, []string) {
	var (
		hostStr   string
		out       bytes.Buffer
		err       error
		fatalList []string
		recapList []string
	)
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
	if err != nil {
		fatalList, recapList = ProcessingOutput(raw)
	}

	return raw, err, fatalList, recapList
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
