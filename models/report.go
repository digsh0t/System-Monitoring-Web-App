package models

import (
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Jeffail/gabs"
	"github.com/wintltr/login-api/auth"
	"github.com/wintltr/login-api/utils"
)

type Report struct {
	Linux_os_total         int            `json:"linux_os_total"`
	Windows_os_total       int            `json:"windows_os_total"`
	Netowrk_os_total       int            `json:"network_os_total"`
	Unknown_os_total       int            `json:"unknown_os_total"`
	SshConnection_total    int            `json:"sshConnection_total"`
	SshKey_total           int            `json:"sshKey_total"`
	Template_total         int            `json:"template_total"`
	EventWeb_total         int            `json:"eventWeb_total"`
	User_total             int            `json:"user_total"`
	CurrentUserWebApp      string         `json:"currentUserWebApp"`
	CurrentUserWebAppRole  string         `json:"currentUserWebAppRole"`
	CurrentUserWebServer   string         `json:"currentUserWebServer"`
	WebServerRunningTime   string         `json:"webserverRunningTime"`
	Current_telegram_token string         `json:"current_telegram_token"`
	Client_report          []ClientReport `json:"client_report"`
}

type ClientReport struct {
	Hostname  string `json:"hostname"`
	Cpu       string `json:"cpu"`
	Serial    string `json:"serial"`
	Osversion string `json:"osversion"`
}

func GetReport(r *http.Request, start time.Time) (Report, error) {
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

	// Get number of eventWeb Log
	evenWebs, err := GetAllEventWebFromDB()
	if err != nil {
		return reportInfo, errors.New("fail to get event web")
	}
	reportInfo.EventWeb_total = len(evenWebs)

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

	// Get Web Server Running time
	elapsed := time.Since(start)
	secondDuration := math.Floor(elapsed.Seconds())

	reportInfo.WebServerRunningTime = secondsToHuman(int(secondDuration))

	/*  // Get Client Report
	reportInfo.Client_report, err = GetClientReport()
	if err != nil {
		return reportInfo, errors.New("fail to get client report")
	}
	*/
	return reportInfo, err

}

func plural(count int, singular string) (result string) {
	if (count == 1) || (count == 0) {
		result = strconv.Itoa(count) + " " + singular + " "
	} else {
		result = strconv.Itoa(count) + " " + singular + "s "
	}
	return
}

func secondsToHuman(input int) (result string) {
	years := math.Floor(float64(input) / 60 / 60 / 24 / 7 / 30 / 12)
	seconds := input % (60 * 60 * 24 * 7 * 30 * 12)
	months := math.Floor(float64(seconds) / 60 / 60 / 24 / 7 / 30)
	seconds = input % (60 * 60 * 24 * 7 * 30)
	weeks := math.Floor(float64(seconds) / 60 / 60 / 24 / 7)
	seconds = input % (60 * 60 * 24 * 7)
	days := math.Floor(float64(seconds) / 60 / 60 / 24)
	seconds = input % (60 * 60 * 24)
	hours := math.Floor(float64(seconds) / 60 / 60)
	seconds = input % (60 * 60)
	minutes := math.Floor(float64(seconds) / 60)
	seconds = input % 60

	if years > 0 {
		result = plural(int(years), "year") + plural(int(months), "month") + plural(int(weeks), "week") + plural(int(days), "day") + plural(int(hours), "hour") + plural(int(minutes), "minute") + plural(int(seconds), "second")
	} else if months > 0 {
		result = plural(int(months), "month") + plural(int(weeks), "week") + plural(int(days), "day") + plural(int(hours), "hour") + plural(int(minutes), "minute") + plural(int(seconds), "second")
	} else if weeks > 0 {
		result = plural(int(weeks), "week") + plural(int(days), "day") + plural(int(hours), "hour") + plural(int(minutes), "minute") + plural(int(seconds), "second")
	} else if days > 0 {
		result = plural(int(days), "day") + plural(int(hours), "hour") + plural(int(minutes), "minute") + plural(int(seconds), "second")
	} else if hours > 0 {
		result = plural(int(hours), "hour") + plural(int(minutes), "minute") + plural(int(seconds), "second")
	} else if minutes > 0 {
		result = plural(int(minutes), "minute") + plural(int(seconds), "second")
	} else {
		result = plural(int(seconds), "second")
	}

	return
}

func GetClientReport() ([]ClientReport, error) {
	var (
		clientReportList []ClientReport
		err              error
	)
	sshConnectionList, err := GetAllSSHConnection()
	if err != nil {
		return clientReportList, err
	}

	// Loop connection
	for _, sshConnection := range sshConnectionList {
		var clientReport ClientReport
		clientReport.Hostname = sshConnection.HostNameSSH
		sysInfo, err := GetLatestSysInfo(sshConnection)
		if err != nil {
			return clientReportList, err
		}
		// Get CPU
		clientReport.Cpu = sysInfo.AvgCPU

		// GET Os Version
		if !sshConnection.IsNetwork {
			clientReport.Osversion = sshConnection.OsType
		} else {
			clientReport.Osversion = sshConnection.NetworkOS
		}

		// Get Serial
		clientReport.Serial, _ = GetClientSerial(sshConnection)

		// Append to list
		clientReportList = append(clientReportList, clientReport)

	}

	// Return
	return clientReportList, err
}

func GetClientSerial(sshConnection SshConnectionInfo) (string, error) {
	var (
		serial string
		err    error
	)

	type SerialJson struct {
		Host string `json:"host"`
	}

	serialJson := SerialJson{
		Host: sshConnection.HostNameSSH,
	}
	serialJsonMarshal, err := json.Marshal(serialJson)
	if err != nil {
		return serial, err
	}
	output, err := RunAnsiblePlaybookWithjson("./yamls/client_getserial.yml", string(serialJsonMarshal))
	if err != nil {
		return serial, err
	}
	if strings.Contains(output, "fatal:") {
		return serial, err
	}

	// Get substring from ansible output
	data := utils.ExtractSubString(output, " => ", "PLAY RECAP")

	// Parse Json format
	jsonParsed, err := gabs.ParseJSON([]byte(data))
	if err != nil {
		return serial, err
	}

	// Get Interfaces
	serial = jsonParsed.Search("msg").String()

	return serial, err

}
