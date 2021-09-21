package models

import (
	"encoding/json"
	"errors"
)

type LinuxClientUser struct {
	UID         string `json:"uid"`
	UID_signed  string `json:"uid_signed"`
	GID         string `json:"gid"`
	GID_signed  string `json:"gid_signed"`
	Username    string `json:"username"`
	Description string `json:"description"`
	Directory   string `json:"directory"`
	Shell       string `json:"shell"`
	UUID        string `json:"uuid"`
	Type        string `json:"type"`
}

type LinuxClientUserJson struct {
	SshConnectionIdList []int    `json:"sshConnectionIdList"`
	Host                []string `json:"host"`
	Username            string   `json:"username"`
	UID                 int      `json:"uid"`
	Groups              []string `json:"group"`
	Comment             string   `json:"comment"`
	Shell               string   `json:"shell"`
}

func LinuxClientUserListAll(sshConnectionId int) ([]LinuxClientUser, error) {
	var (
		clientUserList []LinuxClientUser
		result         string
	)
	SshConnectionInfo, err := GetSSHConnectionFromId(sshConnectionId)
	if err != nil {
		return clientUserList, errors.New("fail to get client connection")
	}

	result, err = SshConnectionInfo.RunCommandFromSSHConnectionUseKeys(`osqueryi --json "SELECT * FROM users"`)
	if err != nil {
		return clientUserList, errors.New("fail to get client users")
	}
	err = json.Unmarshal([]byte(result), &clientUserList)
	if err != nil {
		return clientUserList, errors.New("fail to get client users")
	}

	return clientUserList, nil

}

func LinuxClientUserAdd(userJson LinuxClientUserJson) (string, error) {
	var (
		output string
		err    error
	)

	var host []string
	for _, id := range userJson.SshConnectionIdList {
		sshConnection, err := GetSSHConnectionFromId(id)
		if err != nil {
			return output, errors.New("fail to get list connection")
		}
		host = append(host, sshConnection.HostNameSSH)
	}
	userJson.Host = host

	userJsonMarshal, err := json.Marshal(userJson)
	if err != nil {
		return output, errors.New("fail to marshal json")
	}
	output, err = LoadYAMLWithJson("./yamls/linux_client/add_client_user.yml", string(userJsonMarshal))
	if err != nil {
		return output, err
	}
	return output, err

}

func LinuxClientUserRemove(userJson LinuxClientUserJson) (string, error) {
	var (
		output string
		err    error
	)

	var host []string
	for _, id := range userJson.SshConnectionIdList {
		sshConnection, err := GetSSHConnectionFromId(id)
		if err != nil {
			return output, errors.New("fail to get user connection")
		}
		host = append(host, sshConnection.HostNameSSH)
	}
	userJson.Host = host
	userJsonMarshal, err := json.Marshal(userJson)
	if err != nil {
		return output, errors.New("fail to marshal json")
	}
	output, err = LoadYAMLWithJson("./yamls/linux_client/remove_client_user.yml", string(userJsonMarshal))
	if err != nil {
		return output, err
	}
	return output, nil

}
