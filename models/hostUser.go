package models

import (
	"strings"
)

type HostUserInfo struct {
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
