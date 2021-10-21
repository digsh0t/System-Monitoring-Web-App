package models

import (
	"os"
	"regexp"
	"strconv"
	"strings"
)

type Syslog struct {
	SyslogPRI      int
	SyslogFacility int
	Timegenerated  string
	Hostname       string
	ProgramName    string
	ProcessId      int
	Message        string
}

func extractProcessId(input string) (string, int, error) {
	var processId int
	var err error
	var processName string
	input = strings.Trim(input, ":")
	r, _ := regexp.Compile(`\[(.*)\]`)
	tmp := r.FindString(input)
	if tmp != "" {
		tmp = tmp[1 : len(tmp)-1]
	}
	if tmp == "" {
		processId = -1
	} else {
		processId, err = strconv.Atoi(tmp)
		if err != nil {
			return "", -1, err
		}
	}

	processName = strings.Split(input, "[")[0]
	return processName, processId, nil
}

func parseSyslog(rawLogs string) ([]Syslog, error) {
	var processId int
	var programName string
	var err error
	var logs []Syslog
	for _, line := range strings.Split(strings.Trim(rawLogs, "\r\n "), "\n") {
		var log Syslog
		line = strings.Trim(line, "\r\n\t ")
		vars := strings.SplitN(line, ",", 6)
		programName, processId, err = extractProcessId(vars[4])
		if err != nil {
			return nil, err
		}
		log.SyslogPRI, err = strconv.Atoi(vars[0])
		if err != nil {
			return nil, err
		}
		log.SyslogFacility, err = strconv.Atoi(vars[1])
		if err != nil {
			return nil, err
		}
		log.Timegenerated = vars[2]
		log.Hostname = vars[3]
		log.ProgramName = programName
		log.ProcessId = processId
		log.Message = vars[5]
		logs = append(logs, log)
	}
	return logs, nil
}

func parseSyslogByPri(rawLogs string, pri int) ([]Syslog, error) {
	var processId int
	var programName string
	var err error
	var logs []Syslog
	for _, line := range strings.Split(strings.Trim(rawLogs, "\r\n "), "\n") {
		var log Syslog
		line = strings.Trim(line, "\r\n\t ")
		vars := strings.SplitN(line, ",", 6)
		programName, processId, err = extractProcessId(vars[4])
		if err != nil {
			return nil, err
		}
		log.SyslogPRI, err = strconv.Atoi(vars[0])
		if err != nil {
			return nil, err
		}
		if log.SyslogPRI != pri {
			log = Syslog{}
			continue
		}
		log.SyslogFacility, err = strconv.Atoi(vars[1])
		if err != nil {
			return nil, err
		}
		log.Timegenerated = vars[2]
		log.Hostname = vars[3]
		log.ProgramName = programName
		log.ProcessId = processId
		log.Message = vars[5]
		logs = append(logs, log)
	}
	return logs, nil
}

func GetClientSyslog(logBasePath string, sshConnectionId int, date string) ([]Syslog, error) {
	sshConnection, err := GetSSHConnectionFromId(sshConnectionId)
	if err != nil {
		return nil, err
	}
	sshConnection.HostSSH = "192.168.163.139"
	logPath := logBasePath + "/" + sshConnection.HostSSH + "/" + date + ".log"
	dat, err := os.ReadFile(logPath)
	if err != nil {
		return nil, err
	}
	logs, err := parseSyslog(string(dat))
	if err != nil {
		return nil, err
	}
	return logs, nil
}

func GetClientSyslogByPri(logBasePath string, sshConnectionId int, date string, pri int) ([]Syslog, error) {
	sshConnection, err := GetSSHConnectionFromId(sshConnectionId)
	if err != nil {
		return nil, err
	}
	sshConnection.HostSSH = "192.168.163.139"
	logPath := logBasePath + "/" + sshConnection.HostSSH + "/" + date + ".log"
	dat, err := os.ReadFile(logPath)
	if err != nil {
		return nil, err
	}
	logs, err := parseSyslogByPri(string(dat), pri)
	if err != nil {
		return nil, err
	}
	return logs, nil
}
