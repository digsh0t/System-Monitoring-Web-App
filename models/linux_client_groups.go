package models

import (
	"encoding/json"
	"errors"
)

type LinuxClientGroup struct {
	GID        string `json:"gid"`
	GID_signed string `json:"gid_signed"`
	Groupname  string `json:"groupname"`
}

type LinuxClientGroupJson struct {
	SshConnectionIdList []int    `json:"sshConnectionId"`
	Host                []string `json:"host"`
	Groupname           string   `json:"groupname"`
}

func LinuxClientGroupListAll(hostList []int) ([]LinuxClientGroup, error) {
	var (
		clientGroupList []LinuxClientGroup
		err             error
	)

	// Display installed package on one host
	if len(hostList) == 1 {
		clientGroupList, err = LinuxClientGroupListOfOneHost(hostList[0])
		if err != nil {
			return clientGroupList, err
		}

	} else if len(hostList) > 1 {
		// Display common group on many hosts
		for index, _ := range hostList {
			m := make(map[string]bool)
			var groupList1 []LinuxClientGroup
			var groupList2 []LinuxClientGroup
			if index == 0 {
				groupList1, err = LinuxClientGroupListOfOneHost(hostList[index])
				if err != nil {
					return clientGroupList, err
				}
			} else {
				groupList1 = clientGroupList
				clientGroupList = []LinuxClientGroup{}
			}
			groupList2, err = LinuxClientGroupListOfOneHost(hostList[index+1])
			if err != nil {
				return clientGroupList, err
			}

			for _, groups := range groupList1 {
				m[groups.Groupname] = true
			}

			for _, groups := range groupList2 {
				if _, ok := m[groups.Groupname]; ok {
					clientGroupList = append(clientGroupList, LinuxClientGroup{Groupname: groups.Groupname})
				}
			}
			if len(hostList) == index+2 {
				break
			}

		}
	}
	return clientGroupList, err
}

func LinuxClientGroupListOfOneHost(sshConnectionId int) ([]LinuxClientGroup, error) {
	var (
		clientGroupList []LinuxClientGroup
		result          string
	)
	SshConnectionInfo, err := GetSSHConnectionFromId(sshConnectionId)
	if err != nil {
		return clientGroupList, errors.New("fail to get client connection")
	}

	result, err = SshConnectionInfo.RunCommandFromSSHConnectionUseKeys(`osqueryi --json "SELECT * FROM groups"`)
	if err != nil {
		return clientGroupList, errors.New("fail to get client users")
	}
	err = json.Unmarshal([]byte(result), &clientGroupList)
	if err != nil {
		return clientGroupList, errors.New("fail to get client users")
	}

	return clientGroupList, nil

}

func LinuxClientGroupAdd(groupJson LinuxClientGroupJson) (string, error) {
	var (
		output string
		err    error
	)

	var host []string
	for _, id := range groupJson.SshConnectionIdList {
		sshConnection, err := GetSSHConnectionFromId(id)
		if err != nil {
			return output, errors.New("fail to get list connection")
		}
		host = append(host, sshConnection.HostNameSSH)
	}
	groupJson.Host = host

	groupJsonMarshal, err := json.Marshal(groupJson)
	if err != nil {
		return output, errors.New("fail to marshal json")
	}
	output, err = RunAnsiblePlaybookWithjson("./yamls/linux_client/add_client_group.yml", string(groupJsonMarshal))
	if err != nil {
		return output, err
	}
	return output, err

}

func LinuxClientGroupRemove(groupJson LinuxClientGroupJson) (string, error) {
	var (
		output string
		err    error
	)

	var host []string
	for _, id := range groupJson.SshConnectionIdList {
		sshConnection, err := GetSSHConnectionFromId(id)
		if err != nil {
			return output, errors.New("fail to get user connection")
		}
		host = append(host, sshConnection.HostNameSSH)
	}
	groupJson.Host = host
	groupJsonMarshal, err := json.Marshal(groupJson)
	if err != nil {
		return output, errors.New("fail to marshal json")
	}
	output, err = RunAnsiblePlaybookWithjson("./yamls/linux_client/remove_client_group.yml", string(groupJsonMarshal))
	if err != nil {
		return output, err
	}
	return output, nil

}
