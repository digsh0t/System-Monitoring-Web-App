package models

import "encoding/json"

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
