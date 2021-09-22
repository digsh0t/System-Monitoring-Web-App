package models

import (
	"encoding/json"
)

type localGroup struct {
	Host        interface{} `json:"host"`
	Name        string      `json:"group_name"`
	Description string      `json:"description"`
}

func (sshConnection *SshConnectionInfo) GetLocalUserGroup() ([]LocalUserGroup, error) {
	result, err := sshConnection.RunCommandFromSSHConnectionUseKeys(`osqueryi --json "SELECT * FROM groups"`)
	if err != nil {
		return nil, err
	}
	var groupList []LocalUserGroup

	err = json.Unmarshal([]byte(result), &groupList)
	return groupList, err
}

func AddNewWindowsGroup(host interface{}, name string, description string) (string, error) {
	jsonArgs, err := json.Marshal(localGroup{Host: host, Name: name, Description: description})
	if err != nil {
		return "", err
	}
	output, err := RunAnsiblePlaybookWithjson("yamls/windows_client/add_local_group.yml", string(jsonArgs))
	return output, err
}

func RemoveWindowsGroup(host interface{}, name string) (string, error) {

	type deleteGroup struct {
		Host interface{} `json:"host"`
		Name string      `json:"group_name"`
	}

	jsonArgs, err := json.Marshal(deleteGroup{Host: host, Name: name})
	if err != nil {
		return "", err
	}
	output, err := RunAnsiblePlaybookWithjson("yamls/windows_client/delete_local_group.yml", string(jsonArgs))
	return output, err
}
