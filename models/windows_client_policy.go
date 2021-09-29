package models

import (
	"encoding/json"
	"strings"
)

type registryKey struct {
	Data string `json:"data"`
	Path string `json:"path"`
}

type keys struct {
	Keys []registryKey `json:"key"`
}

func parseKeyList(output string) (keys, error) {
	var keyList keys
	err := json.Unmarshal([]byte(output), &keyList.Keys)
	return keyList, err
}

func (sshConnection SshConnectionInfo) GetExplorerPoliciesSettings(sid string) (keys, error) {
	var regKeyList keys
	result, err := sshConnection.RunCommandFromSSHConnectionUseKeys(`osqueryi --json "SELECT data, path FROM registry WHERE key = 'HKEY_USERS\` + sid + `\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\Explorer'"`)
	if err != nil {
		return regKeyList, err
	}
	regKeyList, err = parseKeyList(result)
	regKeyList.beautifyRegistryKeyList()
	return regKeyList, err
}

func (regKeyList *keys) beautifyRegistryKeyList() {

	pathTranslator := map[string]string{
		"NoControlPanel":     "Disables all Control Panel programs and the PC settings app.",
		"NoDriveTypeAutoRun": "Turn off the Autoplay feature.",
	}

	for i, key := range regKeyList.Keys {
		if strings.Contains(key.Path, "NoControlPanel") {
			regKeyList.Keys[i].Path = pathTranslator["NoControlPanel"]
		}
		if strings.Contains(key.Path, "NoDriveTypeAutoRun") {
			regKeyList.Keys[i].Path = pathTranslator["NoDriveTypeAutoRun"]
		}
	}
}
