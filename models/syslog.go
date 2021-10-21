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

type SyslogPriStat struct {
	Pri0  int `json:"pri_0"`
	Pri1  int `json:"pri_1"`
	Pri2  int `json:"pri_2"`
	Pri3  int `json:"pri_3"`
	Pri4  int `json:"pri_4"`
	Pri5  int `json:"pri_5"`
	Pri6  int `json:"pri_6"`
	Pri7  int `json:"pri_7"`
	Total int `json:"total"`
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

func parseSyslogRowNumbers(rawLogs string) (SyslogPriStat, error) {
	var err error
	var syslogPriStat SyslogPriStat
	for _, line := range strings.Split(strings.Trim(rawLogs, "\r\n "), "\n") {
		var log Syslog
		line = strings.Trim(line, "\r\n\t ")
		vars := strings.SplitN(line, ",", 6)
		log.SyslogPRI, err = strconv.Atoi(vars[0])
		if err != nil {
			return SyslogPriStat{}, err
		}
		switch log.SyslogPRI {
		case 0:
			syslogPriStat.Pri0 += 1
		case 1:
			syslogPriStat.Pri1 += 1
		case 2:
			syslogPriStat.Pri2 += 1
		case 3:
			syslogPriStat.Pri3 += 1
		case 4:
			syslogPriStat.Pri4 += 1
		case 5:
			syslogPriStat.Pri5 += 1
		case 6:
			syslogPriStat.Pri6 += 1
		case 7:
			syslogPriStat.Pri7 += 1
		}
		syslogPriStat.Total += 1
	}
	return syslogPriStat, nil
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

func GetTotalSyslogRows(logBasePath string, sshConnectionId int, date string) ([]Syslog, error) {
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

func GetClientSyslogPriStat(logBasePath string, sshConnectionId int, date string) (SyslogPriStat, error) {
	sshConnection, err := GetSSHConnectionFromId(sshConnectionId)
	if err != nil {
		return SyslogPriStat{}, err
	}
	sshConnection.HostSSH = "192.168.163.139"
	logPath := logBasePath + "/" + sshConnection.HostSSH + "/" + date + ".log"
	dat, err := os.ReadFile(logPath)
	if err != nil {
		return SyslogPriStat{}, err
	}
	logNumbers, err := parseSyslogRowNumbers(string(dat))
	if err != nil {
		return SyslogPriStat{}, err
	}
	return logNumbers, nil
}

func GetAllClientSyslogPriStat(logBasePath string, date string) (SyslogPriStat, error) {
	var syslogPriStat, tmpStat SyslogPriStat
	var err error
	sshConnectionList, err := GetAllSSHConnection()
	if err != nil {
		return SyslogPriStat{}, err
	}
	for _, sshConnection := range sshConnectionList {
		tmpStat, err = GetClientSyslogPriStat("/var/log/remotelogs", sshConnection.SSHConnectionId, date)
		if err != nil {
			return SyslogPriStat{}, err
		}
		syslogPriStat.Pri0 += tmpStat.Pri0
		syslogPriStat.Pri1 += tmpStat.Pri1
		syslogPriStat.Pri2 += tmpStat.Pri2
		syslogPriStat.Pri3 += tmpStat.Pri3
		syslogPriStat.Pri4 += tmpStat.Pri4
		syslogPriStat.Pri5 += tmpStat.Pri5
		syslogPriStat.Pri6 += tmpStat.Pri6
		syslogPriStat.Pri7 += tmpStat.Pri7
		syslogPriStat.Total += tmpStat.Total
		tmpStat = SyslogPriStat{}
	}
	return syslogPriStat, err
}
