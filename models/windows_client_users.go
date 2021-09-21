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

type LocalUserGroup struct {
	Comment   string `json:"comment"`
	Gid       string `json:"gid"`
	GidSigned string `json:"gid_signed"`
	GroupSid  string `json:"group_sid"`
	Groupname string `json:"groupname"`
}

type NewLocalUser struct {
	Host                     []string `json:"host"`
	AccountDisabled          string   `json:"account_disabled"`
	Description              string   `json:"description"`
	Fullname                 string   `json:"fullname"`
	Group                    []string `json:"group"`
	HomeDirectory            string   `json:"home_directory"`
	LoginScript              string   `json:"login_script"`
	Username                 string   `json:"username"`
	Password                 string   `json:"password"`
	PasswordExpired          string   `json:"password_expired"`
	PasswordNeverExpires     string   `json:"password_never_expires"`
	Profile                  string   `json:"profile"`
	UserCannotChangePassword string   `json:"user_cannot_change_password"`
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

func AddNewWindowsUser(userJson string) error {
	err := RunAnsiblePlaybookWithjson(userJson, "./yamls/windows_client/add_local_user.yml")
	return err
}

func DeleteWindowsUser(userJson string) error {
	err := RunAnsiblePlaybookWithjson(userJson, "./yamls/windows_client/delete_local_user.yml")
	return err
}
