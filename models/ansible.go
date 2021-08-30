package models

import (
	"bytes"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type AnsibleInfo struct {
	ExtraValue map[string]string
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

// Load Yaml File
func (ansible *AnsibleInfo) Load(filepath string) (string, error) {
	var (
		out bytes.Buffer
		err error
	)

	// Establish command for load package
	command := "ansible-playbook " + filepath + " -e \""
	for k, v := range ansible.ExtraValue {
		command += k + "=" + v + " "
	}
	command += "\""
	cmd := exec.Command("/bin/bash", "-c", command)
	cmd.Stdout = &out
	err = cmd.Run()
	ouput := out.String()
	return ouput, err
}

// RegExp Fatal And Recap from Ansible Output
func (ansible *AnsibleInfo) ProcessingOutput(raw string) ([]string, []string) {
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

// Convert Recap To Struct Format
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

// Get Hostname From Id And Concentrate its
func (ansible *AnsibleInfo) ProcessingHost(hosts []string) (string, error) {
	var (
		hostStr           string
		sshConnectionList []SshConnectionInfo
		err               error
	)
	// Processing a list host
	if hosts[0] == "all" {
		sshConnectionList, err = GetAllSSHConnection()
		if err != nil {
			return "", err
		}
		for _, v := range sshConnectionList {
			hostStr += v.HostNameSSH + ","
		}
	} else {
		for _, id := range hosts {
			sshConnectionId, err := strconv.Atoi(id)
			sshConnection, err := GetSSHConnectionFromId(sshConnectionId)
			if err != nil {
				return "", err
			}
			hostStr += sshConnection.HostNameSSH + ","
		}
	}
	return hostStr, err
}
