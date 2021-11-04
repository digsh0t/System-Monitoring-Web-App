package models

import (
	"encoding/json"
	"strings"
)

type osProfile struct {
	OS            string         `json:"os"`
	OSKey         windowsLicense `json:"os_key"`
	Manufacturer  string         `json:"manufacturer"`
	Model         string         `json:"model"`
	SerialNumber  string         `json:"serial_number"`
	Processor     string         `json:"processor"`
	OSInstallDate string         `json:"os_install_date"`
}

type physicalDrive struct {
	Name         string `json:"name"`
	Serial       string `json:"serial"`
	Manufacturer string `json:"manufacturer"`
	Model        string `json:"hardware_model"`
	Description  string `json:"description"`
	Type         string `json:"type"`
	Partition    string `json:"partitions"`
	DiskSize     string `json:"disk_size"`
}

type logicalDrive struct {
	Name        string `json:"device_id"`
	Description string `json:"description"`
	Type        string `json:"type"`
	FileSystem  string `json:"file_system"`
	Size        string `json:"size"`
	FreeSpace   string `json:"free_space"`
}

type windowsDefenderStatus struct {
	AMEngineVersion                 string
	AMProductVersion                string
	AMServiceEnabled                bool
	AntispywareEnabled              bool
	AntispywareSignatureLastUpdated string
	AntispywareSignatureVersion     string
	AntivirusEnabled                bool
	AntivirusSignatureLastUpdated   string
	AntivirusSignatureVersion       string
	BehaviorMonitorEnabled          bool
	ComputerState                   string
	FullScanAge                     string
	IoavProtectionEnabled           bool
	IsTamperProtected               bool
	NISEnabled                      bool
	NISEngineVersion                string
	NISSignatureLastUpdated         string
	OnAccessProtectionEnabled       bool
	LastQuickScan                   string
	RealTimeProtectionEnabled       bool
}

func (sshConnection SshConnectionInfo) GetWindowsPhysicalDiskInfo() ([]physicalDrive, error) {
	var driveList []physicalDrive
	result, err := sshConnection.RunCommandFromSSHConnectionUseKeys(`osqueryi --json "SELECT * FROM disk_info";`)
	if err != nil {
		return driveList, err
	}
	err = json.Unmarshal([]byte(result), &driveList)
	if err != nil {
		return driveList, err
	}
	return driveList, err
}

func (sshConnection SshConnectionInfo) GetWindowsLogicalDriveInfo() ([]logicalDrive, error) {
	var logicalDriveList []logicalDrive
	result, err := sshConnection.RunCommandFromSSHConnectionUseKeys(`osqueryi --json "SELECT * FROM logical_drives";`)
	if err != nil {
		return logicalDriveList, err
	}
	err = json.Unmarshal([]byte(result), &logicalDriveList)
	if err != nil {
		return logicalDriveList, err
	}
	return logicalDriveList, err
}

func (sshConnection SshConnectionInfo) GetWindowsDefenderInfo() (windowsDefenderStatus, error) {
	var defenderStatus windowsDefenderStatus
	result, err := sshConnection.RunCommandFromSSHConnectionUseKeys(`powershell -Command "Get-MpComputerStatus | select AMEngineVersion,AMProductVersion,AMServiceEnabled,AntispywareEnabled,AntispywareSignatureLastUpdated,AntispywareSignatureVersion,AntivirusEnabled,AntivirusSignatureLastUpdated,AntivirusSignatureVersion,BehaviorMonitorEnabled,ComputerState,FullScanAge,IoavProtectionEnabled,IsTamperProtected,NISEnabled,NISEngineVersion,NISSignatureLastUpdated,OnAccessProtectionEnabled,QuickScanStartTime,RealTimeProtectionEnabled"`)
	if err != nil {
		return defenderStatus, err
	}
	defenderStatus = parseWindowsDefenderInfoOutput(result)
	return defenderStatus, err
}

func parseWindowsDefenderInfoOutput(input string) windowsDefenderStatus {
	var defenderStatus windowsDefenderStatus
	lines := strings.Split(strings.Trim(input, "\r\n\t"), "\n")
	for i := 0; i < len(lines); i++ {
		lines[i] = strings.Trim(strings.Split(lines[i], ":")[1], "\r\n\t ")
	}
	defenderStatus.AMEngineVersion = lines[0]
	defenderStatus.AMProductVersion = lines[1]
	if lines[2] == "True" {
		defenderStatus.AMServiceEnabled = true
	}
	if lines[3] == "True" {
		defenderStatus.AntispywareEnabled = true
	}
	defenderStatus.AntispywareSignatureLastUpdated = lines[4]
	defenderStatus.AntispywareSignatureVersion = lines[5]
	if lines[6] == "True" {
		defenderStatus.AntivirusEnabled = true
	}
	defenderStatus.AntivirusSignatureLastUpdated = lines[7]
	defenderStatus.AntivirusSignatureVersion = lines[8]
	if lines[9] == "True" {
		defenderStatus.BehaviorMonitorEnabled = true
	}
	switch lines[10] {
	case "0":
		defenderStatus.ComputerState = "clean"
	case "1":
		defenderStatus.ComputerState = "pending full scan"
	case "2":
		defenderStatus.ComputerState = "pending reboot"
	case "4":
		defenderStatus.ComputerState = "pending manual steps"
	case "8":
		defenderStatus.ComputerState = "pending offline scan"
	case "16":
		defenderStatus.ComputerState = "pending critical failure"
	}
	defenderStatus.FullScanAge = lines[11]
	if lines[12] == "True" {
		defenderStatus.IoavProtectionEnabled = true
	}
	if lines[13] == "True" {
		defenderStatus.IsTamperProtected = true
	}
	if lines[14] == "True" {
		defenderStatus.NISEnabled = true
	}
	defenderStatus.NISEngineVersion = lines[15]
	defenderStatus.NISSignatureLastUpdated = lines[16]
	if lines[17] == "True" {
		defenderStatus.OnAccessProtectionEnabled = true
	}
	defenderStatus.LastQuickScan = lines[18]
	if lines[19] == "True" {
		defenderStatus.RealTimeProtectionEnabled = true
	}
	return defenderStatus
}
