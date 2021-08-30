package models

import (
	"strings"
)

type HostUserInfo struct {
	SshConnectionId []string `json:"sshConnectionId"`
	HostUserName    string   `json:"hostUserName"`
	HostUserComment string   `json:"hostUserComment"`
	HostUserUID     string   `json:"hostUserUID"`
	HostUserGroup   string   `json:"hostUserGroup"`
}

func HostUserListAll(sshConnectionId int) ([]string, error) {
	var (
		users  []string
		err    error
		result string
	)
	SshConnectionInfo, err := GetSSHConnectionFromId(sshConnectionId)
	if err != nil {
		return users, err
	}
	command := "getent passwd {1000..60000} | awk -F: '{ print $1}'"
	result, err = RunCommandFromSSHConnection(*SshConnectionInfo, command)
	if err != nil {
		return users, err
	}
	users = strings.Split(strings.TrimSpace(result), "\n")

	return users, nil

}

func (hostUser *HostUserInfo) HostUserAdd() (string, error) {
	var (
		ansible AnsibleInfo
		output  string
	)

	host, err := ansible.ConvertListIdToHostname(hostUser.SshConnectionId)
	if err != nil {
		return output, err
	}
	ansible.ExtraValue = map[string]string{"host": host, "name": hostUser.HostUserName, "comment": hostUser.HostUserComment, "uid": hostUser.HostUserUID, "group": hostUser.HostUserGroup}
	output, err = ansible.Load("./yamls/add_host_user.yml")
	if err != nil {
		return output, err
	}
	return output, nil

}

func (hostUser *HostUserInfo) HostUserRemove() (string, error) {
	var (
		ansible AnsibleInfo
		output  string
	)

	host, err := ansible.ConvertListIdToHostname(hostUser.SshConnectionId)
	if err != nil {
		return output, err
	}
	ansible.ExtraValue = map[string]string{"host": host, "name": hostUser.HostUserName}
	output, err = ansible.Load("./yamls/remove_host_user.yml")
	if err != nil {
		return output, err
	}
	return output, nil

}
