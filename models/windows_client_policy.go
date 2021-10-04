package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type RegistryKey struct {
	Data string `json:"data"`
	Path string `json:"path"`
}

type PasswordPolicy struct {
	ForceLogOff int `json:"force_log_off"`
	MinPwdLen   int `json:"min_pwd_len"`
	MaxPwdAge   int `json:"max_pwd_age"`
	MinPwdAge   int `json:"min_pwd_age"`
	UniquePwd   int `json:"unique_pwd"`
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

func (sshConnection *SshConnectionInfo) UpdateExplorerPolicySettings(uuid string, keyList []RegistryKey) error {
	uglifyRegistryKeyList(keyList)
	type modifyRegistryKeyList struct {
		Host         string        `json:"host"`
		RegistryPath string        `json:"registry_path"`
		Key          []RegistryKey `json:"key"`
		DataType     string        `json:"data_type"`
	}
	var registryKeyList modifyRegistryKeyList
	registryKeyList.Host = sshConnection.HostNameSSH
	userBasePath, err := sshConnection.regLoadCurrentUser(uuid)
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

func (sshConnection *SshConnectionInfo) UpdateWindowsUserProhibitedProgramsPolicy(uuid string, programList []string) error {
	type modifyRegistry struct {
		Host         string        `json:"host"`
		RegistryPath string        `json:"registry_path"`
		Key          []RegistryKey `json:"key"`
		DataType     string        `json:"data_type"`
	}
	var registry modifyRegistry
	var keyList []RegistryKey

	registry.Host = sshConnection.HostNameSSH
	userBasePath, err := sshConnection.regLoadCurrentUser(uuid)
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

func (sshConnection SshConnectionInfo) regLoadCurrentUser(uuid string) (string, error) {
	username, err := sshConnection.getWindowsUsernameFromUUID(uuid)
	if err != nil {
		return "", err
	}
	path := `C:\users\` + username + `\ntuser.dat`
	_, err = sshConnection.RunCommandFromSSHConnectionUseKeys(`reg load HKU\` + username + " " + path)
	if err != nil {
		if strings.Contains(err.Error(), "The process cannot access the file because it is being used by another process.") {
			return `HKU:\` + uuid, nil
		}

		return "", err
	}

	return `HKU:\` + username, nil
}

func (sshConnection SshConnectionInfo) getWindowsUsernameFromUUID(uuid string) (string, error) {
	command := fmt.Sprintf(`osqueryi --json "SELECT username FROM users WHERE uuid LIKE '%s'"`, uuid)
	output, err := sshConnection.RunCommandFromSSHConnectionUseKeys(command)
	if err != nil {
		return "", err
	}
	var userList []User
	json.Unmarshal([]byte(output), &userList)
	return userList[0].Username, nil
}

func (sshConnection SshConnectionInfo) GetWindowsPasswordPolicy() (PasswordPolicy, error) {
	output, err := sshConnection.RunCommandFromSSHConnectionUseKeys(`net accounts`)
	if err != nil {
		return PasswordPolicy{}, err
	}
	policy, err := parsePasswordPolicyResult(output)
	return policy, err
}

func parsePasswordPolicyResult(output string) (PasswordPolicy, error) {
	var policy PasswordPolicy
	var values []string
	var err error
	for i, line := range strings.Split(output, "\n") {
		if i > 4 {
			break
		}
		keyAndValue := strings.Split(line, ":")
		keyAndValue[1] = strings.Trim(keyAndValue[1], "\r\n ")
		if strings.Contains(keyAndValue[1], "Never") || strings.Contains(keyAndValue[1], "Unlimited") || strings.Contains(keyAndValue[1], "None") {
			keyAndValue[1] = "0"
		}
		values = append(values, keyAndValue[1])
	}
	policy.ForceLogOff, err = strconv.Atoi(values[0])
	if err != nil {
		return policy, err
	}
	policy.MinPwdAge, err = strconv.Atoi(values[1])
	if err != nil {
		return policy, err
	}
	policy.MaxPwdAge, err = strconv.Atoi(values[2])
	if err != nil {
		return policy, err
	}
	policy.MinPwdLen, err = strconv.Atoi(values[3])
	if err != nil {
		return policy, err
	}
	policy.UniquePwd, err = strconv.Atoi(values[4])
	return policy, err

}

func (sshConnection SshConnectionInfo) ChangeWindowsPasswordPolicy(policy PasswordPolicy) error {
	var forceLogOff, maxPwdAge, minPwdLen, minPwdAge, uniquePwd string
	if policy.ForceLogOff == 0 {
		forceLogOff = "NO"
	} else {
		forceLogOff = strconv.Itoa(policy.ForceLogOff)
	}
	if policy.MaxPwdAge == 0 {
		maxPwdAge = "UNLIMITED"
	} else {
		maxPwdAge = strconv.Itoa(policy.MaxPwdAge)
	}
	minPwdAge = strconv.Itoa(policy.MinPwdAge)
	minPwdLen = strconv.Itoa(policy.MinPwdLen)
	uniquePwd = strconv.Itoa(policy.UniquePwd)
	command := fmt.Sprintf(`net accounts /FORCELOGOFF:%s /MINPWLEN:%s /MAXPWAGE:%s /MINPWAGE:%s /UNIQUEPW:%s`, forceLogOff, minPwdLen, maxPwdAge, minPwdAge, uniquePwd)
	output, err := sshConnection.RunCommandFromSSHConnectionUseKeys(command)
	if !strings.Contains(output, "The command completed successfully") {
		return errors.New(output)
	}
	return err
}
