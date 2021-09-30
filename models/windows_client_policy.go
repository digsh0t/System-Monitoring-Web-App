package models

import (
	"encoding/json"
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
	result, err := sshConnection.RunCommandFromSSHConnectionUseKeys(`osqueryi --json "SELECT data, path FROM registry WHERE key = 'HKEY_USERS\` + sid + `\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\Explorer'"`)
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
	}

	for i, key := range regKeyList {
		if strings.Contains(key.Path, "NoControlPanel") {
			regKeyList[i].Path = pathTranslator["NoControlPanel"]
		}
		if strings.Contains(key.Path, "NoDriveTypeAutoRun") {
			regKeyList[i].Path = pathTranslator["NoDriveTypeAutoRun"]
		}
	}
}

func uglifyRegistryKeyList(regKeyList []RegistryKey) {

	pathTranslator := map[string]string{
		"Disables all Control Panel programs and the PC settings app": "NoControlPanel",
		"Turn off the Autoplay feature":                               "NoDriveTypeAutoRun",
	}

	for i, key := range regKeyList {
		if strings.Contains(key.Path, "Disables all Control Panel programs and the PC settings app") {
			regKeyList[i].Path = pathTranslator["Disables all Control Panel programs and the PC settings app"]
		}
		if strings.Contains(key.Path, "Turn off the Autoplay feature") {
			regKeyList[i].Path = pathTranslator["Turn off the Autoplay feature"]
		}
	}
}

func (sshConnection *SshConnectionInfo) UpdateExplorerPolicySettings(sid string, keyList []RegistryKey) error {
	uglifyRegistryKeyList(keyList)
	type modifyRegistryKeyList struct {
		Host         string        `json:"host"`
		RegistryPath string        `json:"registry_path"`
		Key          []RegistryKey `json:"key"`
	}
	var registryKeyList modifyRegistryKeyList
	registryKeyList.Host = sshConnection.HostNameSSH
	registryKeyList.RegistryPath = `HKU:\` + sid + `\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\Explorer`
	registryKeyList.Key = keyList
	marshalled, err := json.Marshal(registryKeyList)
	if err != nil {
		return err
	}
	_, err = RunAnsiblePlaybookWithjson("./yamls/windows_client/add_or_update_registry.yml", string(marshalled))
	return err
}
