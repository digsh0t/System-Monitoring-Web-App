package models

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type SysInfo struct {
	ConnectionId int    `json:"id"`
	HostnameSSH  string `json:"hostnameSSH"`
	AvgCPU       string `json:"avgcpu"`
	AvgMem       string `json:"avgmem"`
	Timestamp    string `json:"timestamp"`
}

func GetSysInfo(sshConnection SshConnectionInfo) (SysInfo, error) {
	var err error
	var sysInfo SysInfo
	sysInfo.ConnectionId = sshConnection.SSHConnectionId
	sysInfo.HostnameSSH = sshConnection.HostNameSSH
	sysInfo.AvgMem, err = CalcAvgMemUse(sshConnection)
	if err != nil {
		return sysInfo, err
	}
	sysInfo.AvgCPU, err = CalcAvgCPUFromTop(sshConnection)
	sysInfo.Timestamp = time.Now().Format("01-02-2006 15:04:05")
	return sysInfo, err
}

func CalcAvgCPUFromTop(sshConnection SshConnectionInfo) (string, error) {
	var cpuUse float32

	command := "top -b -n 1"
	result, err := RunCommandFromSSHConnection(sshConnection, command)
	if err != nil {
		return "", err
	}

	lines := strings.Split(result, "\n")
	for _, line := range lines {
		if strings.Contains(line, "%Cpu(s):") {
			atributes := strings.Split(line, ",")
			idle, err := strconv.ParseFloat(strings.Trim((atributes[3][:5]), " "), 32)
			if err != nil {
				return "", err
			}
			cpuUse = 100 - float32(idle)
		}
	}
	return fmt.Sprintf("%.1f", cpuUse), nil
}

func CalcAvgMemUse(sshConnection SshConnectionInfo) (string, error) {
	result, err := ExecCommand("free | grep Mem | awk '{print $3/$2 * 100.0}'", sshConnection.UserSSH, sshConnection.PasswordSSH, sshConnection.HostSSH, sshConnection.PortSSH)
	return strings.Trim(result, "\n"), err
}

func GetAllSysInfo(sshConnectionList []SshConnectionInfo) ([]SysInfo, error) {
	var sysInfoList []SysInfo
	var sysInfo SysInfo
	var err error

	for _, sshConnection := range sshConnectionList {
		sshConnection.PasswordSSH = AESDecryptKey(sshConnection.PasswordSSH)
		sysInfo, err = GetSysInfo(sshConnection)
		if err != nil {
			return sysInfoList, err
		}
		sysInfoList = append(sysInfoList, sysInfo)
	}

	return sysInfoList, nil
}
