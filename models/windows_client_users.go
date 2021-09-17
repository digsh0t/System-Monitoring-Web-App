package models

import "encoding/json"

type ClientUser struct {
	Description string `json:"description"`
	Directory   string `json:"directory"`
	Gid         string `json:"gid"`
	GidSigned   string `json:"gid_signed"`
	Shell       string `json:"shell"`
	Type        string `json:"type"`
	UID         string `json:"uid"`
	UIDSigned   string `json:"uid_signed"`
	Username    string `json:"username"`
	UUID        string `json:"uuid"`
}

//Both Linux and Windows can use this
func (sshConnection *SshConnectionInfo) GetLocalUsers() ([]ClientUser, error) {
	result, err := sshConnection.RunCommandFromSSHConnectionUseKeys(`osqueryi --json "SELECT * FROM users"`)
	if err != nil {
		return nil, err
	}
	var userList []ClientUser

	err = json.Unmarshal([]byte(result), &userList)
	return userList, err
}
