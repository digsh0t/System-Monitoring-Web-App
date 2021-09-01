package models

import (
	"strings"
)

type HostUserInfo struct {
	SshConnectionIdList []string `json:"sshConnectionIdList"`
	SshConnectionId     int      `json:"sshConnectionId"`
	HostUserName        string   `json:"hostUserName"`
	HostUserComment     string   `json:"hostUserComment"`
	HostUserUID         string   `json:"hostUserUID"`
	HostUserGroup       string   `json:"hostUserGroup"`
	HostUserPassword    string   `json:"hostuserPassword"`
	HostUserHomeDir     string   `json:"hostuserHomeDir"`
	HostUserShell       string   `json:"hostuserShell"`
}

func HostUserListAll(sshConnectionId int) ([]HostUserInfo, error) {
	var (
		hostUserList []HostUserInfo
		result       string
	)
	SshConnectionInfo, err := GetSSHConnectionFromId(sshConnectionId)
	if err != nil {
		return hostUserList, err
	}
	command := "getent passwd {1000..60000}"
	result, err = RunCommandFromSSHConnection(*SshConnectionInfo, command)
	if !strings.Contains(err.Error(), "Process exited with status 2") {
		return hostUserList, err
	}
	lines := strings.Split(strings.TrimSpace(result), "\n")
	for _, line := range lines {
		var hostuser HostUserInfo
		attributesUser := strings.Split(line, ":")
		hostuser.SshConnectionId = sshConnectionId
		hostuser.HostUserName = attributesUser[0]
		hostuser.HostUserPassword = attributesUser[1]
		hostuser.HostUserUID = attributesUser[2]
		gid := attributesUser[3]
		hostuser.HostUserComment = attributesUser[4]
		hostuser.HostUserHomeDir = attributesUser[5]
		hostuser.HostUserShell = attributesUser[6]
		command = "getent group " + gid + " | cut -d: -f1"
		groupname, err := RunCommandFromSSHConnection(*SshConnectionInfo, command)
		if err != nil {
			return hostUserList, err
		}
		hostuser.HostUserGroup = strings.TrimSpace(groupname)
		hostUserList = append(hostUserList, hostuser)
	}

	return hostUserList, nil

}

func (hostUser *HostUserInfo) HostUserAdd() (string, error) {
	var (
		ansible AnsibleInfo
		output  string
	)

	host, err := ansible.ConvertListIdToHostname(hostUser.SshConnectionIdList)
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

	host, err := ansible.ConvertListIdToHostname(hostUser.SshConnectionIdList)
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
