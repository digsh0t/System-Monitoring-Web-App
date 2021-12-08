package models

import (
	"encoding/json"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
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

func InstallWindowsProgram(host interface{}, url string, dest string) (string, error) {

	type installInfo struct {
		Host     interface{} `json:"host"`
		Url      string      `json:"url"`
		Dest     string      `json:"dest"`
		Filename string      `json:"filename"`
	}
	filename := filepath.Base(url)
	dest = `C:\tmp` + strconv.Itoa(rand.Int()) + `\`
	jsonArgs, err := json.Marshal(installInfo{Host: host, Url: url, Dest: dest, Filename: filename})
	if err != nil {
		return "", err
	}
	output, err := RunAnsiblePlaybookWithjson("yamls/windows_client/install_windows_program.yml", string(jsonArgs))
	return output, err
}

func DeleteWindowsProgram(host interface{}, productId string) (string, error) {

	type deleteInfo struct {
		Host      interface{} `json:"host"`
		ProductId string      `json:"product_id"`
	}

	jsonArgs, err := json.Marshal(deleteInfo{Host: host, ProductId: productId})
	if err != nil {
		return "", err
	}
	output, err := RunAnsiblePlaybookWithjson("yamls/windows_client/delete_windows_program.yml", string(jsonArgs))
	return output, err
}

func InstallWindowsProgramLocal(host interface{}, filename string) (string, error) {

	type installInfo struct {
		Host     interface{} `json:"host"`
		Src      string      `json:"source"`
		Dest     string      `json:"dest"`
		Filename string      `json:"filename"`
	}

	source, err := os.Getwd()
	if err != nil {
		return "", err
	}
	source += "/msi"
	dest := `C:\tmp` + strconv.Itoa(rand.Int()) + `\`
	jsonArgs, err := json.Marshal(installInfo{Host: host, Src: source, Dest: dest, Filename: filename})
	if err != nil {
		return "", err
	}
	output, err := RunAnsiblePlaybookWithjson("yamls/windows_client/install_windows_program_local_msi.yml", string(jsonArgs))
	return output, err
}
