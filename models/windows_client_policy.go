package models

import (
	"encoding/json"
	"strconv"
	"strings"
)

type RegistryKey struct {
	Data string `json:"data"`
	Path string `json:"path"`
}

func parseKeyList(output string) ([]RegistryKey, error) {
	var keyList []RegistryKey
	err := json.Unmarshal([]byte(output), &keyList)
	return keyList, err
}

func (sshConnection SshConnectionInfo) GetExplorerPoliciesSettings(sid string) ([]RegistryKey, error) {
	var regKeyList []RegistryKey
	result, err := sshConnection.RunCommandFromSSHConnectionUseKeys(`osqueryi --json "SELECT data, path FROM registry WHERE key = 'HKEY_USERS\` + sid + `\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\Explorer' AND data != ''"`)
	if err != nil {
		return regKeyList, err
	}
	regKeyList, err = parseKeyList(result)
	beautifyRegistryKeyList(regKeyList)
	return regKeyList, err
}

func beautifyRegistryKeyList(regKeyList []RegistryKey) {

	pathTranslator := map[string]string{
		"NoControlPanel":     "Disables all Control Panel programs and the PC settings app.",
		"NoDriveTypeAutoRun": "Turn off the Autoplay feature.",
		"DisallowRun":        "Prevent Users From Running Certain Programs",
	}

	for i, key := range regKeyList {
		if strings.Contains(key.Path, "NoControlPanel") {
			regKeyList[i].Path = pathTranslator["NoControlPanel"]
		}
		if strings.Contains(key.Path, "NoDriveTypeAutoRun") {
			regKeyList[i].Path = pathTranslator["NoDriveTypeAutoRun"]
		}
		if strings.Contains(key.Path, "DisallowRun") {
			regKeyList[i].Path = pathTranslator["DisallowRun"]
		}
	}
}

func uglifyRegistryKeyList(regKeyList []RegistryKey) {

	pathTranslator := map[string]string{
		"Disables all Control Panel programs and the PC settings app": "NoControlPanel",
		"Turn off the Autoplay feature":                               "NoDriveTypeAutoRun",
		"Prevent Users From Running Certain Programs":                 "DisallowRun",
	}

	for i, key := range regKeyList {
		if strings.Contains(key.Path, "Disables all Control Panel programs and the PC settings app") {
			regKeyList[i].Path = pathTranslator["Disables all Control Panel programs and the PC settings app"]
		}
		if strings.Contains(key.Path, "Turn off the Autoplay feature") {
			regKeyList[i].Path = pathTranslator["Turn off the Autoplay feature"]
		}
		if strings.Contains(key.Path, "Prevent Users From Running Certain Programs") {
			regKeyList[i].Path = pathTranslator["Prevent Users From Running Certain Programs"]
		}
	}
}

func (sshConnection *SshConnectionInfo) UpdateExplorerPolicySettings(sid string, keyList []RegistryKey) error {
	uglifyRegistryKeyList(keyList)
	type modifyRegistryKeyList struct {
		Host         string        `json:"host"`
		RegistryPath string        `json:"registry_path"`
		Key          []RegistryKey `json:"key"`
		DataType     string        `json:"data_type"`
	}
	var registryKeyList modifyRegistryKeyList
	registryKeyList.Host = sshConnection.HostNameSSH
	userBasePath, err := sshConnection.regLoadCurrentUser(sid)
	if err != nil {
		return err
	}
	path := userBasePath + `\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\Explorer`
	registryKeyList.RegistryPath = path
	registryKeyList.Key = keyList
	registryKeyList.DataType = "dword"

	marshalled, err := json.Marshal(registryKeyList)
	if err != nil {
		return err
	}
	_, err = RunAnsiblePlaybookWithjson("./yamls/windows_client/add_or_update_registry.yml", string(marshalled))
	if err != nil {
		return err
	}
	err = sshConnection.unloadReg()
	return err
}

func (sshConnection *SshConnectionInfo) GetProhibitedProgramsPolicy(sid string) ([]string, error) {
	var regKeyList []RegistryKey
	var programList []string
	result, err := sshConnection.RunCommandFromSSHConnectionUseKeys(`osqueryi --json "SELECT data, path FROM registry WHERE key = 'HKEY_USERS\` + sid + `\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\Explorer\DisallowRun' AND data != ''"`)
	if err != nil {
		return nil, err
	}
	regKeyList, err = parseKeyList(result)
	for _, key := range regKeyList {
		programList = append(programList, key.Data)
	}
	return programList, err
}

func (sshConnection *SshConnectionInfo) UpdateWindowsUserProhibitedProgramsPolicy(sid string, programList []string) error {
	type modifyRegistry struct {
		Host         string        `json:"host"`
		RegistryPath string        `json:"registry_path"`
		Key          []RegistryKey `json:"key"`
		DataType     string        `json:"data_type"`
	}
	var registry modifyRegistry
	var keyList []RegistryKey

	registry.Host = sshConnection.HostNameSSH
	userBasePath, err := sshConnection.regLoadCurrentUser("wintltr")
	if err != nil {
		return err
	}
	registry.RegistryPath = userBasePath + `\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\Explorer\DisallowRun`
	for i, program := range programList {
		keyList = append(keyList, RegistryKey{Data: program, Path: strconv.Itoa(i + 1)})
	}
	registry.Key = keyList
	registry.DataType = "string"
	marshalled, err := json.Marshal(registry)
	if err != nil {
		return err
	}
	_, err = RunAnsiblePlaybookWithjson("./yamls/windows_client/add_or_update_registry.yml", string(marshalled))
	return err
}

func (sshConnection SshConnectionInfo) regLoadCurrentUser(username string) (string, error) {
	path := `C:\users\` + username + `\ntuser.dat`
	_, err := sshConnection.RunCommandFromSSHConnectionUseKeys(`reg load HKU\CurrentUser ` + path)
	if err != nil {
		if strings.Contains(err.Error(), "The process cannot access the file because it is being used by another process.") {
			return `HKCU:`, nil
		}

		return "", err
	}

	return `HKU:\CurrentUser`, nil
}

func (sshConnection SshConnectionInfo) unloadReg() error {
	_, err := sshConnection.RunCommandFromSSHConnectionUseKeys(`reg unload HKU\CurrentUser`)
	return err
}
