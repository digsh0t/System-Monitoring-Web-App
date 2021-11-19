package models

import (
	"encoding/json"
	"strconv"
)

type process struct {
	Cmdline              string `json:"cmdline"`
	Cwd                  string `json:"cwd"`
	DiskBytesRead        string `json:"disk_bytes_read"`
	DiskBytesWritten     string `json:"disk_bytes_written"`
	Egid                 string `json:"egid"`
	ElapsedTime          string `json:"elapsed_time"`
	ElevatedToken        string `json:"elevated_token"`
	Euid                 string `json:"euid"`
	Gid                  string `json:"gid"`
	HandleCount          string `json:"handle_count"`
	Name                 string `json:"name"`
	Nice                 string `json:"nice"`
	OnDisk               string `json:"on_disk"`
	Parent               string `json:"parent"`
	Path                 string `json:"path"`
	PercentProcessorTime string `json:"percent_processor_time"`
	Pgroup               string `json:"pgroup"`
	Pid                  string `json:"pid"`
	ProtectionType       string `json:"protection_type"`
	ResidentSize         string `json:"resident_size"`
	Root                 string `json:"root"`
	SecureProcess        string `json:"secure_process"`
	Sgid                 string `json:"sgid"`
	StartTime            string `json:"start_time"`
	State                string `json:"state"`
	Suid                 string `json:"suid"`
	SystemTime           string `json:"system_time"`
	Threads              string `json:"threads"`
	TotalSize            string `json:"total_size"`
	UID                  string `json:"uid"`
	UserTime             string `json:"user_time"`
	Username             string `json:"username"`
	VirtualProcess       string `json:"virtual_process"`
	WiredSize            string `json:"wired_size"`
}

func (sshConnection SshConnectionInfo) GetProcessListFromWindows() ([]process, error) {
	result, err := sshConnection.RunCommandFromSSHConnectionUseKeys(`osqueryi --json "SELECT P.*,U.username FROM processes AS P LEFT JOIN users AS U WHERE P.uid=U.uid"`)
	if err != nil {
		return nil, err
	}
	processList, err := parseProcessFromCmd(result)
	return processList, err
}

func parseProcessFromCmd(input string) ([]process, error) {
	var processList []process
	err := json.Unmarshal([]byte(input), &processList)
	return processList, err
}

func (sshConnection SshConnectionInfo) KillProcess(pid int) (string, error) {
	command := "taskkill /f /t /pid " + strconv.Itoa(pid)
	result, err := sshConnection.RunCommandFromSSHConnectionUseKeys(command)
	return result, err
}
