package models

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
)

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

func AddNewWindowsUser(userJson string) (string, error) {
	output, err := RunAnsiblePlaybookWithjson("./yamls/windows_client/add_local_user.yml", userJson)

	return output, err
}

func DeleteWindowsUser(userJson string) (string, error) {
	output, err := RunAnsiblePlaybookWithjson("./yamls/windows_client/delete_local_user.yml", userJson)
	return output, err
}

func (sshConnection *SshConnectionInfo) GetWindowsGroupUserBelongTo(username string) ([]string, error) {
	isValid, err := regexp.MatchString("^[a-zA-Z0-9]+$", username)
	if !isValid || err != nil {
		return nil, err
	}
	result, err := sshConnection.RunCommandFromSSHConnectionUseKeys(`osqueryi --json "SELECT G.groupname FROM user_groups UG INNER JOIN users U ON U.uid=UG.uid INNER JOIN groups G ON G.gid=UG.GID WHERE U.username='` + username + `'`)
	if err != nil {
		return nil, err
	}
	type groupName struct {
		Groupname string `json:"groupname"`
	}
	var gNL []groupName
	var strGroupNameList []string
	err = json.Unmarshal([]byte(result), &gNL)
	for _, groupName := range gNL {
		strGroupNameList = append(strGroupNameList, groupName.Groupname)
	}
	return strGroupNameList, err
}

func (sshConnectionInfo *SshConnectionInfo) ReplaceWindowsGroupForUser(username string, group []string) (string, error) {
	type replacedGroup struct {
		Host     string   `json:"host"`
		Username string   `json:"username"`
		Group    []string `json:"group"`
	}

	groupListJson, err := json.Marshal(replacedGroup{Host: sshConnectionInfo.HostNameSSH, Username: username, Group: group})
	if err != nil {
		return "", err
	}
	output, err := RunAnsiblePlaybookWithjson("./yamls/windows_client/change_user_group_membership.yml", string(groupListJson))
	return output, err
}

func parseLoggedInUser(output string) ([]loggedInUser, error) {
	var loggedInUserList []loggedInUser

	var user loggedInUser
	re := regexp.MustCompile(`\s{2,}`)
	for i, line := range strings.Split(strings.Trim(output, "\n\r "), "\n") {
		if i == 0 {
			continue
		}
		line = strings.Trim(line, "\n\r ")
		vars := re.Split(line, -1)
		if len(vars) == 6 {
			user.Username = vars[0]
			user.SessionName = vars[1]
			user.SessionId = vars[2]
			user.State = vars[3]
			if vars[3] == "Disc" {
				user.State = "Disconnected"
			}
			user.IdleTime = vars[4]
			user.LogonTime = vars[5]
		} else if len(vars) == 5 {
			user.Username = vars[0]
			user.SessionId = vars[1]
			user.State = vars[2]
			if vars[2] == "Disc" {
				user.State = "Disconnected"
			}
			user.IdleTime = vars[3]
			user.LogonTime = vars[4]
		}
		loggedInUserList = append(loggedInUserList, user)
		user = loggedInUser{}
	}

	return loggedInUserList, nil
}

func (sshConnection SshConnectionInfo) GetLoggedInUsers() ([]loggedInUser, error) {
	var loggedInUserList []loggedInUser
	result, err := sshConnection.RunCommandFromSSHConnectionUseKeys(`quser`)
	if err != nil {
		return loggedInUserList, err
	}
	loggedInUserList, err = parseLoggedInUser(result)
	return loggedInUserList, err
}

func (sshConnection SshConnectionInfo) KillWindowsLoginSession(sessionId int) error {
	_, err := sshConnection.RunCommandFromSSHConnectionUseKeys(`logoff ` + strconv.Itoa(sessionId))
	if err != nil {
		return err
	}
	return err
}
