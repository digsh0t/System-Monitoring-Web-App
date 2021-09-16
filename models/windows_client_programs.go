package models

import (
	"encoding/json"
	"path/filepath"
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

func InstallWindowsProgram(host []string, url string, dest string) error {

	type installInfo struct {
		Host     []string `json:"host"`
		Url      string   `json:"url"`
		Dest     string   `json:"dest"`
		Filename string   `json:"filename"`
	}

	filename := filepath.Base(url)
	jsonArgs, err := json.Marshal(installInfo{Host: host, Url: url, Dest: dest, Filename: filename})
	if err != nil {
		return err
	}
	err = RunAnsiblePlaybookWithjson(string(jsonArgs), "yamls/windows_client/add_windows_program.yml")
	return err
}

func DeleteWindowsProgram(host []string, productId string) error {

	type deleteInfo struct {
		Host      []string `json:"host"`
		ProductId string   `json:"product_id"`
	}

	jsonArgs, err := json.Marshal(deleteInfo{Host: host, ProductId: productId})
	if err != nil {
		return err
	}
	err = RunAnsiblePlaybookWithjson(string(jsonArgs), "yamls/windows_client/delete_windows_program.yml")
	return err
}
