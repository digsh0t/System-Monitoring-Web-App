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

func InstallWindowsProgram(hostUrlDesFilenameJson string) error {

	err := RunAnsiblePlaybookWithjson(hostUrlDesFilenameJson, "yamls/windows_client/add_windows_program.yml")
	return err
}

func DeleteWindowsProgram(host interface{}, productId string) error {

	type deleteInfo struct {
		Host      interface{} `json:"host"`
		ProductId string      `json:"product_id"`
	}

	jsonArgs, err := json.Marshal(deleteInfo{Host: host, ProductId: productId})
	if err != nil {
		return err
	}
	err = RunAnsiblePlaybookWithjson(string(jsonArgs), "yamls/windows_client/delete_windows_program.yml")
	return err
}
