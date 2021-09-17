package models

import (
	"bytes"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/Jeffail/gabs"
	"github.com/wintltr/login-api/utils"
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
func LoadYAML(filepath string, extraValue map[string]string) (string, error) {
	var (
		out    bytes.Buffer
		err    error
		output string
	)

	// Establish command for load package
	command := "ansible-playbook " + filepath + " -e \""
	for k, v := range extraValue {
		command += k + "=" + v + " "
	}
	command += "\""
	cmd := exec.Command("/bin/bash", "-c", command)
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		// "Exit status 2" means Ansible displays fatal error but our funtion still works correctly
		if err.Error() == "exit status 2" || err.Error() == "exit status 4" {
			err = nil
		} else {
			return output, err
		}
	}
	output = out.String()
	return output, err
}

// RegExp Fatal And Recap from Ansible Output
func ProcessingAnsibleOutput(ansible_output string) (map[string]bool, []string, error) {
	var (
		status    map[string]bool
		fatalList []string
		recapList []string
		err       error
	)

	// Extracting Fatal
	text := strings.Split(ansible_output, "\n")
	for _, line := range text {
		pattern := "^fatal"
		r, _ := regexp.Compile(pattern)
		if r.MatchString(line) {
			msg, err := ParseFatal(line)
			if err != nil {
				return status, fatalList, err
			}
			fatalList = append(fatalList, msg)
		}
	}

	// Extracting PLAY RECAP **********
	pattern := "PLAY RECAP .+\n"
	r, _ := regexp.Compile(pattern)
	strIndex := r.FindStringIndex(ansible_output)
	tmp := ansible_output[strIndex[1]:]
	text = strings.Split(tmp, "\n")
	for _, line := range text {
		if line != "" {
			recapList = append(recapList, line)
		}
	}
	recapInfoList, err := ParseRecap(recapList)
	if err != nil {
		return status, fatalList, err
	}
	status = AnalyzeRecap(recapInfoList)

	return status, fatalList, err

}

func ParseFatal(line string) (string, error) {
	var (
		msg string
		err error
	)
	data := utils.ExtractSubStringByStartIndex(line, " => ")
	jsonParsed, err := gabs.ParseJSON([]byte(data))
	if err != nil {
		return msg, err
	}
	msg = jsonParsed.Search("msg").String()
	return msg, err
}

// Convert Recap To Struct Format
func ParseRecap(recapList []string) ([]RecapInfo, error) {
	var (
		recapInfo     RecapInfo
		recapInfoList []RecapInfo
		err           error
	)

	for _, line := range recapList {
		pattern := "^(.+)\\s+:"
		r, _ := regexp.Compile(pattern)
		stringSubmatch := r.FindStringSubmatch(line)
		recapInfo.ClientName = strings.TrimSpace(stringSubmatch[1])

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
				recapInfo.Ok = number
			case "changed":
				recapInfo.Changed = number
			case "unreachable":
				recapInfo.Unreachable = number
			case "failed":
				recapInfo.Failed = number
			case "skipped":
				recapInfo.Skipped = number
			case "rescued":
				recapInfo.Rescued = number
			case "ignored":
				recapInfo.Ignored = number
			}
		}
		recapInfoList = append(recapInfoList, recapInfo)

	}
	return recapInfoList, err
}

func AnalyzeRecap(RecapInfoList []RecapInfo) map[string]bool {
	result := make(map[string]bool)
	for _, recapInfo := range RecapInfoList {
		if recapInfo.Failed > 0 || recapInfo.Unreachable > 0 {
			result[recapInfo.ClientName] = false
		} else {
			result[recapInfo.ClientName] = true
		}
	}
	return result
}

// Get Hostname From Ids type []string And Concentrate its
func ConvertListIdToHostname(hosts []string) (string, error) {
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

// Get Hostname From Ids type []int And Concentrate its
func ConvertListIdToHostnameVer2(hosts []int) (string, error) {
	var (
		hostStr string
		err     error
	)
	for _, id := range hosts {
		sshConnection, err := GetSSHConnectionFromId(id)
		if err != nil {
			return "", err
		}
		hostStr += sshConnection.HostNameSSH + ","
	}

	return hostStr, err
}
