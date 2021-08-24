package models

import (
	"bytes"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type YamlInfo struct {
	Host    []string `json:"host"`
	File    string   `json:"file"`
	Mode    string   `json:"mode"`
	Package string   `json:"package"`
	Link    string   `json:"link"`
}

type RecapInfo struct {
	ClientName  string
	Ok          int
	Changed     int
	Unreachable int
	Failed      int
	Skipped     int
	Rescued     int
	Ignored     int
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

func (recapStruct *RecapInfo) ProcessingRecap(recapList []string) ([]RecapInfo, error) {
	var recapStructList []RecapInfo
	var err error
	for _, line := range recapList {
		pattern := "^(.+)\\s+:"
		r, _ := regexp.Compile(pattern)
		stringSubmatch := r.FindStringSubmatch(line)
		recapStruct.ClientName = stringSubmatch[1]

		for _, keyword := range []string{"ok", "changed", "unreachable", "failed", "skipped", "rescued", "ignored"} {
			pattern = keyword + "=" + "([0-9]+)"
			r, _ := regexp.Compile(pattern)
			stringSubmatch = r.FindStringSubmatch(line)
			number, err := strconv.Atoi(stringSubmatch[1])
			if err != nil {
				return nil, err
			}
			switch keyword {
			case "ok":
				recapStruct.Ok = number
			case "changed":
				recapStruct.Changed = number
			case "unreachable":
				recapStruct.Unreachable = number
			case "failed":
				recapStruct.Failed = number
			case "skipped":
				recapStruct.Skipped = number
			case "rescued":
				recapStruct.Rescued = number
			case "ignored":
				recapStruct.Ignored = number
			}
		}
		recapStructList = append(recapStructList, *recapStruct)

	}
	return recapStructList, err
}
