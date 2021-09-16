package models

import (
	"encoding/json"
)

type Programs struct {
	Name              string `json:"name"`
	Version           string `json:"version"`
	InstallLocation   string `json:"install_location"`
	InstalSource      string `json:"install_source"`
	Language          string `json:"language"`
	Publisher         string `json:"publisher"`
	UninstallString   string `json:"uninstall_string"`
	InstallDate       string `json:"install_date"`
	IdentifyingNumber string `json:"identifying_number"`
}

func GetInstalledProgram(sshConnection SshConnectionInfo) ([]Programs, error) {
	result, err := sshConnection.RunCommandFromSSHConnectionUseKeys(`osqueryi --json "SELECT * FROM programs"`)
	if err != nil {
		return nil, err
	}
	var installedPrograms []Programs

	err = json.Unmarshal([]byte(result), &installedPrograms)
	return installedPrograms, err
}
