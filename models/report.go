package models

import (
	"errors"
	"net/http"

	"github.com/wintltr/login-api/auth"
	"github.com/wintltr/login-api/utils"
)

type Report struct {
	Linux_os_total         int    `json:"linux_os_total"`
	Windows_os_total       int    `json:"windows_os_total"`
	Netowrk_os_total       int    `json:"network_os_total"`
	Unknown_os_total       int    `json:"unknown_os_total"`
	SshConnection_total    int    `json:"sshConnection_total"`
	SshKey_total           int    `json:"sshKey_total"`
	Template_total         int    `json:"template_total"`
	User_total             int    `json:"user_total"`
	CurrentUserWebApp      string `json:"currentUserWebApp"`
	CurrentUserWebAppRole  string `json:"currentUserWebAppRole"`
	CurrentUserWebServer   string `json:"currentUserWebServer"`
	Current_telegram_token string `json:"current_telegram_token"`
}

func GetReport(r *http.Request) (Report, error) {
	var (
		reportInfo Report
		err        error
	)

	// Get Linux Os total
	sshConnectionLinux, err := GetAllOSSSHConnection("Linux")
	if err != nil {
		return reportInfo, errors.New("fail to get linux os total")
	}
	reportInfo.Linux_os_total = len(sshConnectionLinux)

	// Get Windows Os total
	sshConnectionWindows, err := GetAllOSSSHConnection("Windows")
	if err != nil {
		return reportInfo, errors.New("fail to get windows os total")
	}
	reportInfo.Windows_os_total = len(sshConnectionWindows)

	// Get Unknown Os
	reportInfo.Unknown_os_total, err = CountUnknownOS()
	if err != nil {
		return reportInfo, errors.New("fail to get unknown os total")
	}

	// Get Network Os
	reportInfo.Netowrk_os_total, err = CountNetworkOS()
	if err != nil {
		return reportInfo, errors.New("fail to get network os total")
	}

	// Get Total ssh Connection
	reportInfo.SshConnection_total = reportInfo.Linux_os_total + reportInfo.Netowrk_os_total + reportInfo.Windows_os_total + reportInfo.Unknown_os_total

	// Get SshKey total
	sshKeys, err := GetAllSSHKeyFromDB()
	if err != nil {
		return reportInfo, errors.New("fail to get ssh key total")
	}
	reportInfo.SshKey_total = len(sshKeys)

	// Get Template total
	templates, err := GetAllTemplate()
	if err != nil {
		return reportInfo, errors.New("fail to get template total")
	}
	reportInfo.Template_total = len(templates)

	// Get User total
	users, err := GetAllUserFromDB()
	if err != nil {
		return reportInfo, errors.New("fail to get user total")
	}
	reportInfo.User_total = len(users)

	// Get Current User Web app
	tokenData, err := auth.RetrieveTokenData(r)
	if err != nil {
		return reportInfo, errors.New("fail to get current user web app")
	}
	reportInfo.CurrentUserWebApp = tokenData.Username

	// Get Current User Web app role
	reportInfo.CurrentUserWebAppRole = tokenData.Role

	// Get Current User Web Server
	reportInfo.CurrentUserWebServer = utils.GetCurrentUser().Name

	// Get Current telegram token
	apikey, err := GetTelegramAPIKey()
	if err != nil {
		return reportInfo, errors.New("fail to get current telegram key")
	}
	reportInfo.Current_telegram_token = apikey.ApiToken

	return reportInfo, err
}
