package models

import (
	"regexp"
	"strings"
)

type lastLogin struct {
	Username      string `json:"username"`
	LastLoginTime string `josn:"last_login_time"`
}

func (sshConnection SshConnectionInfo) GetLinuxUsersLastLogin() ([]lastLogin, error) {
	var lastLoginList []lastLogin
	output, err := sshConnection.RunCommandFromSSHConnectionUseKeys("lastlog")
	if err != nil {
		return lastLoginList, err
	}
	lastLoginList, err = parseLinuxLastlog(output)
	return lastLoginList, err

}

func parseLinuxLastlog(input string) ([]lastLogin, error) {
	var lastLoginList []lastLogin
	var lastLoginInfo lastLogin
	re, err := regexp.Compile(`\s{3,}`)
	if err != nil {
		return lastLoginList, err
	}
	for i, line := range strings.Split(strings.Trim(input, "\r\n\t "), "\n") {
		if i == 0 {
			continue
		}
		vars := re.Split(line, -1)
		lastLoginInfo.Username = vars[0]
		if strings.Contains(vars[len(vars)-1], "**Never logged in**") {
			lastLoginInfo.LastLoginTime = "Never"
		} else {
			lastLoginInfo.LastLoginTime = vars[len(vars)-1]
		}
		lastLoginList = append(lastLoginList, lastLoginInfo)
		lastLoginInfo = lastLogin{}
	}
	return lastLoginList, err
}
