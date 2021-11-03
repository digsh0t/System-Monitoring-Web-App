package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Jeffail/gabs"
	"github.com/jung-kurt/gofpdf"
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

	reportInfo.WebServerRunningTime = utils.SecondsToHuman(int(secondDuration))

	/*  // Get Client Report
	reportInfo.Client_report, err = GetClientReport()
	if err != nil {
		return reportInfo, errors.New("fail to get client report")
	}
	*/
	return reportInfo, err

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

func ExportReport(filename string) error {

	// Cover Page
	type recType struct {
		align, txt string
	}

	recList := []recType{
		{"CM", "Web Application Report"},
		{"BC", utils.GetCurrentDateTime()},
	}

	var formatRect = func(pdf *gofpdf.Fpdf, recList []recType) {
		pdf.AddPage()
		pdf.SetMargins(10, 10, 10)
		pdf.SetAutoPageBreak(false, 0)
		borderStr := "1"
		pdf.CellFormat(190, 257, "Version 1.0", "", 1, "BC", false, 0, "")

		pdf.ImageOptions(
			"./pictures/fpt.png",
			25, 70,
			0, 0,
			false,
			gofpdf.ImageOptions{ImageType: "PNG", ReadDpi: true},
			0,
			"",
		)
		for _, rec := range recList {
			pdf.SetXY(20, 20)
			pdf.CellFormat(170, 257, rec.txt, borderStr, 0, rec.align, false, 0, "")
			borderStr = ""
		}
	}

	pdf := gofpdf.New("P", "mm", "A4", "") // A4 210.0 x 297.0
	pdf.SetHeaderFuncMode(func() {
		pdf.ImageOptions(
			"./pictures/logo4.png",
			4, 4,
			0, 0,
			false,
			gofpdf.ImageOptions{ImageType: "PNG", ReadDpi: true},
			0,
			"",
		)
		pdf.WriteAligned(190, 4, "Asset Detail Report", "R")
		pdf.Ln(20)
	}, true)
	pdf.SetFooterFunc(func() {
		pdf.SetY(-15)
		pdf.SetFont("Arial", "I", 8)
		pdf.CellFormat(0, 10, fmt.Sprintf("Page %d/{nb}", pdf.PageNo()),
			"", 0, "C", false, 0, "")
	})
	pdf.AliasNbPages("")
	pdf.SetFont("Arial", "B", 16)
	formatRect(pdf, recList)

	// Second Page
	pdf.AddPage()
	pdf.CellFormat(60, 30, "Table of Contents ", "", 0, "C", false, 0, "")
	pdf.Ln(20)
	pdf.WriteAligned(90, 20, "1. Computer devices", "L")
	pdf.Ln(10)
	pdf.WriteAligned(70, 20, "   1.1. Windows devices", "L")
	pdf.Ln(10)

	// Get Windows
	sshConnectionList, err := GetAllOSSSHConnection("Windows")
	if err != nil {
		return err
	}
	for index, sshConnection := range sshConnectionList {
		pdf.WriteAligned(100, 20, "          1.1."+strconv.Itoa(index+1)+". "+sshConnection.HostNameSSH, "L")
	}

	pdf.Ln(10)
	pdf.WriteAligned(70, 20, "   1.2. Linux devices", "L")
	pdf.Ln(10)

	// Get Linux
	sshConnectionList, err = GetAllOSSSHConnection("Linux")
	if err != nil {
		return err
	}
	for index, sshConnection := range sshConnectionList {
		pdf.WriteAligned(100, 20, "          1.2."+strconv.Itoa(index+1)+". "+sshConnection.HostNameSSH, "L")
		pdf.Ln(8)
	}

	pdf.Ln(2)
	pdf.WriteAligned(70, 20, "2. Network devices", "L")
	pdf.Ln(10)
	pdf.WriteAligned(70, 20, "   2.1. Router devices", "L")
	pdf.Ln(10)

	// Get Router
	sshConnectionList, err = GetAllOSSSHConnection("Router")
	if err != nil {
		return err
	}
	for index, sshConnection := range sshConnectionList {
		pdf.WriteAligned(100, 20, "          2.1."+strconv.Itoa(index+1)+". "+sshConnection.HostNameSSH, "L")
		pdf.Ln(8)
	}

	pdf.Ln(2)
	pdf.WriteAligned(70, 20, "   2.2. Switch devices", "L")
	pdf.Ln(10)

	// Get Switch
	sshConnectionList, err = GetAllOSSSHConnection("Switch")
	if err != nil {
		return err
	}
	for index, sshConnection := range sshConnectionList {
		pdf.WriteAligned(100, 20, "          2.2."+strconv.Itoa(index+1)+". "+sshConnection.HostNameSSH, "L")
		pdf.Ln(8)
	}
	pdf.Ln(10)

	// Content

	var DrawTable = func(pdf *gofpdf.Fpdf, sshConnectionList []SshConnectionInfo) error {
		for index, sshConnection := range sshConnectionList {
			pdf.SetFont("Arial", "B", 16)
			if index > 0 {
				pdf.AddPage()
			}
			sshConnectionInfo, err := GetSSHConnectionInformationBySSH_Id(sshConnection.SSHConnectionId)
			if err != nil {
				return err
			}
			pdf.WriteAligned(100, 20, "          1.1."+strconv.Itoa(index+1)+". "+sshConnection.HostNameSSH, "L")
			pdf.SetY(80)
			pdf.CellFormat(180, 7, "Profile", "1", 0, "", false, 0, "")
			pdf.Ln(-1)
			for i := 0; i < 7; i++ {
				pdf.SetFont("", "B", 12)
				switch i {
				case 0:
					pdf.CellFormat(60, 6, "OS Name", "1", 0, "", false, 0, "")
					pdf.CellFormat(120, 6, sshConnectionInfo.OsName, "1", 0, "", false, 0, "")
					pdf.Ln(-1)
				case 1:
					pdf.CellFormat(60, 6, "OS Version", "1", 0, "", false, 0, "")
					pdf.CellFormat(120, 6, sshConnectionInfo.OsVersion, "1", 0, "", false, 0, "")
					pdf.Ln(-1)
				case 2:
					pdf.CellFormat(60, 6, "OS Install Date", "1", 0, "", false, 0, "")
					pdf.CellFormat(120, 6, sshConnectionInfo.InstallDate, "1", 0, "", false, 0, "")
					pdf.Ln(-1)
				case 3:
					pdf.CellFormat(60, 6, "Serial Number", "1", 0, "", false, 0, "")
					pdf.CellFormat(120, 6, sshConnectionInfo.Serial, "1", 0, "", false, 0, "")
					pdf.Ln(-1)
				case 4:
					pdf.CellFormat(60, 6, "OS Host Name", "1", 0, "", false, 0, "")
					pdf.CellFormat(120, 6, sshConnectionInfo.Hostname, "1", 0, "", false, 0, "")
					pdf.Ln(-1)
				case 5:
					pdf.CellFormat(60, 6, "Manufacturer", "1", 0, "", false, 0, "")
					pdf.CellFormat(120, 6, sshConnectionInfo.Manufacturer, "1", 0, "", false, 0, "")
					pdf.Ln(-1)
				case 6:
					pdf.CellFormat(60, 6, "OS Model", "1", 0, "", false, 0, "")
					pdf.CellFormat(120, 6, sshConnectionInfo.Model, "1", 0, "", false, 0, "")
					pdf.Ln(-1)
				case 7:
					pdf.CellFormat(60, 6, "Architecture", "1", 0, "", false, 0, "")
					pdf.CellFormat(120, 6, sshConnectionInfo.Architecture, "1", 0, "", false, 0, "")

				}
			}

		}
		return err
	}
	pdf.AddPage()
	pdf.SetFont("", "B", 15)
	pdf.Ln(10)
	pdf.WriteAligned(90, 20, "Contents", "L")
	pdf.Ln(20)
	pdf.WriteAligned(90, 20, "1. Computer devices", "L")
	pdf.Ln(10)
	pdf.WriteAligned(70, 20, "   1.1. Windows devices", "L")
	pdf.Ln(10)

	// Get Windows
	sshConnectionList, err = GetAllOSSSHConnection("Windows")
	if err != nil {
		return err
	}
	err = DrawTable(pdf, sshConnectionList)
	if err != nil {
		return err
	}

	// Get Linux
	pdf.AddPage()
	pdf.Ln(20)
	pdf.SetFont("", "B", 15)
	pdf.WriteAligned(70, 20, "   1.2. Linux devices", "L")
	pdf.Ln(10)
	sshConnectionList, err = GetAllOSSSHConnection("Linux")
	if err != nil {
		return err
	}
	err = DrawTable(pdf, sshConnectionList)
	if err != nil {
		return err
	}

	// Get Router
	pdf.AddPage()
	pdf.SetFont("", "B", 15)
	pdf.WriteAligned(90, 20, "2. Network devices", "L")
	pdf.Ln(10)
	pdf.WriteAligned(70, 20, "   2.1. Router devices", "L")
	pdf.Ln(10)
	sshConnectionList, err = GetAllOSSSHConnection("Router")
	if err != nil {
		return err
	}
	err = DrawTable(pdf, sshConnectionList)
	if err != nil {
		return err
	}

	// Get Switch
	pdf.AddPage()
	pdf.SetFont("", "B", 15)
	pdf.WriteAligned(70, 20, "   2.2. Switch devices", "L")
	pdf.Ln(10)
	sshConnectionList, err = GetAllOSSSHConnection("Switch")
	if err != nil {
		return err
	}
	err = DrawTable(pdf, sshConnectionList)
	if err != nil {
		return err
	}

	return pdf.OutputFileAndClose(filename)
}
