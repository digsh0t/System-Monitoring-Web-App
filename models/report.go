package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
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
		pdf.CellFormat(280, 360, "Version 1.0", "", 1, "BC", false, 0, "")
		pdf.ImageOptions(
			"./pictures/fpt.png",
			70, 70,
			0, 0,
			false,
			gofpdf.ImageOptions{ImageType: "PNG", ReadDpi: true},
			0,
			"",
		)
		for _, rec := range recList {
			pdf.SetXY(20, 20)
			pdf.CellFormat(258, 365, rec.txt, borderStr, 0, rec.align, false, 0, "")
			borderStr = ""
		}
	}

	// Initialize
	pdf := gofpdf.New("P", "mm", "A3", "") // A4 210.0 x 297.
	pdf.SetAutoPageBreak(true, 20.0)
	pdf.SetHeaderFuncMode(func() {
		pdf.SetFont("", "B", 12)
		pdf.ImageOptions(
			"./pictures/logo4.png",
			4, 4,
			0, 0,
			false,
			gofpdf.ImageOptions{ImageType: "PNG", ReadDpi: true},
			0,
			"",
		)
		pdf.WriteAligned(270, 4, "Asset Detail Report", "R")
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
	var DrawSystemInfoTable = func(pdf *gofpdf.Fpdf, index int, sshConnection SshConnectionInfo) error {
		pdf.SetAutoPageBreak(true, 20.0)

		// Draw System info
		if index > 0 {
			pdf.AddPage()
		}
		pdf.SetFont("Arial", "B", 16)
		sshConnectionInfo, err := GetSSHConnectionInformationBySSH_Id(sshConnection.SSHConnectionId)
		if err != nil {
			return err
		}
		pdf.WriteAligned(100, 20, "          1.1."+strconv.Itoa(index+1)+". "+sshConnection.HostNameSSH, "L")
		pdf.SetY(80)
		pdf.SetFillColor(141, 151, 173)
		pdf.SetTextColor(255, 255, 255)
		pdf.SetLineWidth(.3)
		pdf.SetFont("", "B", 0)
		pdf.CellFormat(180, 7, "Profile", "1", 0, "", true, 0, "")
		pdf.Ln(-1)

		// Color and font restoration
		pdf.SetFillColor(224, 235, 255)
		pdf.SetTextColor(0, 0, 0)
		pdf.SetFont("", "", 0)
		// 	Data
		fill := false
		for i := 0; i < 7; i++ {
			pdf.SetFont("", "B", 12)
			switch i {
			case 0:
				pdf.CellFormat(60, 6, "OS Name", "1", 0, "", fill, 0, "")
				pdf.SetFont("", "", 12)
				pdf.CellFormat(120, 6, sshConnectionInfo.OsName, "1", 0, "", fill, 0, "")
				pdf.Ln(-1)
			case 1:
				pdf.CellFormat(60, 6, "OS Version", "1", 0, "", fill, 0, "")
				pdf.SetFont("", "", 12)
				pdf.CellFormat(120, 6, sshConnectionInfo.OsVersion, "1", 0, "", fill, 0, "")
				pdf.Ln(-1)
			case 2:
				pdf.CellFormat(60, 6, "OS Install Date", "1", 0, "", fill, 0, "")
				pdf.SetFont("", "", 12)
				pdf.CellFormat(120, 6, sshConnectionInfo.InstallDate, "1", 0, "", fill, 0, "")
				pdf.Ln(-1)
			case 3:
				pdf.CellFormat(60, 6, "Serial Number", "1", 0, "", fill, 0, "")
				pdf.SetFont("", "", 12)
				pdf.CellFormat(120, 6, sshConnectionInfo.Serial, "1", 0, "", fill, 0, "")
				pdf.Ln(-1)
			case 4:
				pdf.CellFormat(60, 6, "OS Host Name", "1", 0, "", fill, 0, "")
				pdf.SetFont("", "", 12)
				pdf.CellFormat(120, 6, sshConnectionInfo.Hostname, "1", 0, "", fill, 0, "")
				pdf.Ln(-1)
			case 5:
				pdf.CellFormat(60, 6, "Manufacturer", "1", 0, "", fill, 0, "")
				pdf.SetFont("", "", 12)
				pdf.CellFormat(120, 6, sshConnectionInfo.Manufacturer, "1", 0, "", fill, 0, "")
				pdf.Ln(-1)
			case 6:
				pdf.CellFormat(60, 6, "OS Model", "1", 0, "", fill, 0, "")
				pdf.SetFont("", "", 12)
				pdf.CellFormat(120, 6, sshConnectionInfo.Model, "1", 0, "", fill, 0, "")
				pdf.Ln(-1)
			case 7:
				pdf.CellFormat(60, 6, "Architecture", "1", 0, "", fill, 0, "")
				pdf.SetFont("", "", 12)
				pdf.CellFormat(120, 6, sshConnectionInfo.Architecture, "1", 0, "", fill, 0, "")

			}
			fill = !fill
		}
		return err
	}

	var DrawPhysDriveTable = func(pdf *gofpdf.Fpdf, index int, sshConnection SshConnectionInfo) error {
		pdf.SetAutoPageBreak(true, 20.0)

		// Draw System info
		header := []string{"Name", "Serial", "Manufacturer", "Model", "Description", "Type", "Partition", "DiskSize"}

		pdf.Ln(20)
		pdf.WriteAligned(100, 20, "Windows Physical Disk", "L")
		pdf.Ln(20)
		pdf.SetFillColor(141, 151, 173)
		pdf.SetTextColor(255, 255, 255)
		pdf.SetLineWidth(.3)
		pdf.SetFont("", "B", 0)
		w := []float64{45.0, 45.0, 50.0, 25.0, 25.0, 25.0, 25.0, 30.0}
		for j, str := range header {
			pdf.CellFormat(w[j], 7, str, "1", 0, "C", true, 0, "")
		}
		pdf.Ln(-1)

		// Color and font restoration
		pdf.SetFillColor(224, 235, 255)
		pdf.SetTextColor(0, 0, 0)
		pdf.SetFont("", "", 0)
		// 	Data
		fill := false
		physDiskInfoList, err := sshConnection.GetWindowsPhysicalDiskInfo()
		if err != nil {
			return err
		}

		for _, physDisk := range physDiskInfoList {

			// Get height of Model column
			marginCell := 2.
			_, lineHt := pdf.GetFontSize()
			height := 0.
			lines := pdf.SplitLines([]byte(physDisk.Model), w[3])
			h := float64(len(lines))*lineHt + marginCell*float64(len(lines))
			if h > height {
				height = h
			}

			pdf.CellFormat(w[0], height, physDisk.Name, "1", 0, "", fill, 0, "")
			pdf.CellFormat(w[1], height, physDisk.Serial, "1", 0, "", fill, 0, "")
			pdf.CellFormat(w[2], height, physDisk.Manufacturer, "1", 0, "", fill, 0, "")
			// Get current position
			curx, y := pdf.GetXY()
			pdf.MultiCell(w[3], lineHt+marginCell, physDisk.Model, "1", "", fill)
			// Restore position
			pdf.SetXY(curx+w[3], y)
			pdf.CellFormat(w[4], height, physDisk.Description, "1", 0, "", fill, 0, "")
			pdf.CellFormat(w[5], height, physDisk.Type, "1", 0, "", fill, 0, "")
			pdf.CellFormat(w[6], height, physDisk.Partition, "1", 0, "", fill, 0, "")
			pdf.CellFormat(w[7], height, physDisk.DiskSize, "1", 0, "", fill, 0, "")
			pdf.Ln(-1)
			fill = !fill
		}

		return err
	}

	var DrawLogicDriveTable = func(pdf *gofpdf.Fpdf, index int, sshConnection SshConnectionInfo) error {
		pdf.SetAutoPageBreak(true, 20.0)

		// Draw System info
		header := []string{"Name", "Description", "Type", "FileSystem", "Size", "FreeSpace"}

		pdf.Ln(20)
		pdf.WriteAligned(100, 20, "Windows Logical Disk", "L")
		pdf.Ln(20)
		pdf.SetFillColor(141, 151, 173)
		pdf.SetTextColor(255, 255, 255)
		pdf.SetLineWidth(.3)
		pdf.SetFont("", "B", 0)
		w := []float64{45.0, 45.0, 50.0, 25.0, 45.0, 45.0}
		for j, str := range header {
			pdf.CellFormat(w[j], 7, str, "1", 0, "C", true, 0, "")
		}
		pdf.Ln(-1)

		// Color and font restoration
		pdf.SetFillColor(224, 235, 255)
		pdf.SetTextColor(0, 0, 0)
		pdf.SetFont("", "", 0)
		// 	Data
		fill := false
		logicalDiskInfoList, err := sshConnection.GetWindowsLogicalDriveInfo()
		if err != nil {
			return err
		}

		for _, logicalDisk := range logicalDiskInfoList {
			var height float64 = 6
			pdf.CellFormat(w[0], height, logicalDisk.Name, "1", 0, "", fill, 0, "")
			pdf.CellFormat(w[1], height, logicalDisk.Description, "1", 0, "", fill, 0, "")
			pdf.CellFormat(w[2], height, logicalDisk.Type, "1", 0, "", fill, 0, "")
			pdf.CellFormat(w[3], height, logicalDisk.FileSystem, "1", 0, "", fill, 0, "")
			if logicalDisk.Size == "-1" {
				pdf.CellFormat(w[4], height, "N/A", "1", 0, "", fill, 0, "")
			} else {
				pdf.CellFormat(w[4], height, logicalDisk.Size, "1", 0, "", fill, 0, "")
			}
			if logicalDisk.FreeSpace == "-1" {
				pdf.CellFormat(w[4], height, "N/A", "1", 0, "", fill, 0, "")
			} else {
				pdf.CellFormat(w[4], height, logicalDisk.FreeSpace, "1", 0, "", fill, 0, "")
			}
			pdf.Ln(-1)
			fill = !fill
		}

		return err
	}

	var DrawLocalUserTable = func(pdf *gofpdf.Fpdf, index int, sshConnection SshConnectionInfo) error {
		pdf.SetAutoPageBreak(true, 20.0)
		pdf.Ln(20)
		pdf.WriteAligned(100, 20, "Local Users", "L")
		pdf.Ln(20)
		localUserList, err := sshConnection.GetLocalUsers()
		if err != nil {
			return err
		}

		pdf.SetFillColor(141, 151, 173)
		pdf.SetTextColor(255, 255, 255)
		pdf.SetLineWidth(.3)
		pdf.SetFont("", "B", 0)
		pdf.SetX(45)
		pdf.CellFormat(200, 10, "Users", "1", 0, "C", true, 0, "")
		pdf.Ln(-1)
		// Color and font restoration
		pdf.SetFillColor(224, 235, 255)
		pdf.SetTextColor(0, 0, 0)
		pdf.SetFont("", "B", 0)

		// Draw header
		pdf.SetX(45)
		pdf.CellFormat(80, 10, "local user accounts", "", 0, "C", true, 0, "")
		pdf.CellFormat(120, 10, "last logon", "", 0, "C", true, 0, "")
		pdf.SetFont("", "", 0)
		pdf.Ln(-1)
		for _, localUser := range localUserList {
			pdf.SetX(53)
			if !localUser.IsEnabled {
				pdf.CellFormat(20, 10, "X", "", 0, "", false, 0, "")
			}
			pdf.SetX(62)
			pdf.CellFormat(100, 10, localUser.Username, "", 0, "", false, 0, "")
			pdf.CellFormat(150, 10, localUser.LastLogon, "", 0, "", false, 0, "")
			pdf.Ln(-1)
		}
		pdf.SetX(80)
		pdf.CellFormat(100, 10, "X Marks a disabled account;", "", 0, "", false, 0, "")

		return err
	}

	//var WriteError = func(pdf *gofpdf.Fpdf, index int, sshConnection SshConnectionInfo) {
	//	pdf.CellFormat(100, 10, "Error while connecting "+sshConnection.HostNameSSH+", please check again to see this report information", "", 0, "", false, 0, "")
	//}

	var DrawWindowsInterfaceTable = func(pdf *gofpdf.Fpdf, index int, sshConnection SshConnectionInfo) error {
		pdf.SetAutoPageBreak(true, 20.0)

		// Draw System info
		header := []string{"ConnectionName", "Description", "IP", "Mac", "DHCPServer", "Subnet", "InterfaceType", "Manufacturer", "DefaultGateway", "DNSDomain"}

		pdf.Ln(20)
		pdf.WriteAligned(100, 20, "Windows Interfaces Information", "L")
		pdf.Ln(20)
		pdf.SetFillColor(141, 151, 173)
		pdf.SetTextColor(255, 255, 255)
		pdf.SetLineWidth(.3)
		pdf.SetFont("", "B", 0)
		w := []float64{45.0, 45.0, 50.0, 25.0, 45.0, 45.0, 45.0, 45.0, 45.0, 45.0}
		for j, str := range header {
			pdf.CellFormat(w[j], 7, str, "1", 0, "C", true, 0, "")
		}
		pdf.Ln(-1)

		// Color and font restoration
		pdf.SetFillColor(224, 235, 255)
		pdf.SetTextColor(0, 0, 0)
		pdf.SetFont("", "", 0)
		// 	Data
		fill := false
		interfacesList, err := sshConnection.GetWindowsInterfaceInfo()
		if err != nil {
			return err
		}

		for _, interfaces := range interfacesList {
			var height float64 = 6
			pdf.CellFormat(w[0], height, interfaces.ConnectionName, "1", 0, "", fill, 0, "")
			pdf.CellFormat(w[1], height, interfaces.Description, "1", 0, "", fill, 0, "")
			pdf.CellFormat(w[2], height, interfaces.IP, "1", 0, "", fill, 0, "")
			pdf.CellFormat(w[3], height, interfaces.Mac, "1", 0, "", fill, 0, "")
			pdf.CellFormat(w[4], height, interfaces.DHCPServer, "1", 0, "", fill, 0, "")
			pdf.CellFormat(w[5], height, interfaces.Subnet, "1", 0, "", fill, 0, "")
			pdf.CellFormat(w[6], height, interfaces.InterfaceType, "1", 0, "", fill, 0, "")
			pdf.CellFormat(w[7], height, interfaces.Manufacturer, "1", 0, "", fill, 0, "")
			pdf.CellFormat(w[8], height, interfaces.DefaultGateway, "1", 0, "", fill, 0, "")
			pdf.CellFormat(w[9], height, interfaces.DNSDomain, "1", 0, "", fill, 0, "")
			pdf.Ln(-1)
			fill = !fill
		}

		return err
	}

	/*var DrawLinuxInterfaceTable = func(pdf *gofpdf.Fpdf, index int, sshConnection SshConnectionInfo) error {
		pdf.SetAutoPageBreak(true, 20.0)

		// Draw System info
		header := []string{"Active", "InterfaceName", "IPv4", "IPV6", "Mac", "InterfaceType"}

		pdf.Ln(20)
		pdf.WriteAligned(100, 20, "Windows Interfaces Information", "L")
		pdf.Ln(20)
		pdf.SetFillColor(141, 151, 173)
		pdf.SetTextColor(255, 255, 255)
		pdf.SetLineWidth(.3)
		pdf.SetFont("", "B", 0)
		w := []float64{45.0, 45.0, 50.0, 25.0, 45.0, 45.0}
		for j, str := range header {
			pdf.CellFormat(w[j], 7, str, "1", 0, "C", true, 0, "")
		}
		pdf.Ln(-1)

		// Color and font restoration
		pdf.SetFillColor(224, 235, 255)
		pdf.SetTextColor(0, 0, 0)
		pdf.SetFont("", "", 0)
		// 	Data
		fill := false
		interfacesList, err := sshConnection.GetLinuxInterfaceInfo()
		if err != nil {
			return err
		}

		for _, interfaces := range interfacesList {
			var height float64 = 6
			pdf.CellFormat(w[0], height, strconv.FormatBool(interfaces.Active), "1", 0, "", fill, 0, "")
			pdf.CellFormat(w[1], height, interfaces.InterfaceName, "1", 0, "", fill, 0, "")
			pdf.CellFormat(w[2], height, interfaces.IPv4, "1", 0, "", fill, 0, "")
			pdf.CellFormat(w[3], height, interfaces.Mac, "1", 0, "", fill, 0, "")
			pdf.CellFormat(w[4], height, interfaces.DHCPServer, "1", 0, "", fill, 0, "")
			pdf.CellFormat(w[5], height, interfaces.Subnet, "1", 0, "", fill, 0, "")
			pdf.CellFormat(w[6], height, interfaces.InterfaceType, "1", 0, "", fill, 0, "")
			pdf.CellFormat(w[7], height, interfaces.Manufacturer, "1", 0, "", fill, 0, "")
			pdf.CellFormat(w[8], height, interfaces.DefaultGateway, "1", 0, "", fill, 0, "")
			pdf.CellFormat(w[9], height, interfaces.DNSDomain, "1", 0, "", fill, 0, "")
			pdf.Ln(-1)
			fill = !fill
		}

		return err
	}*/

	/*var DrawWindowsProgramTable = func(pdf *gofpdf.Fpdf, index int, sshConnection SshConnectionInfo) error {
		pdf.SetAutoPageBreak(true, 20.0)
		pdf.Ln(20)
		pdf.WriteAligned(100, 20, "Windows Program", "L")
		pdf.Ln(20)
		programList, err := sshConnection.GetInstalledProgram()
		if err != nil {
			return err
		}

		pdf.SetFillColor(141, 151, 173)
		pdf.SetTextColor(255, 255, 255)
		pdf.SetLineWidth(.3)
		pdf.SetFont("", "B", 0)
		pdf.SetX(45)
		pdf.CellFormat(200, 10, "Software Versions and Usage", "1", 0, "C", true, 0, "")
		pdf.Ln(-1)
		// Color and font restoration
		pdf.SetFillColor(224, 235, 255)
		pdf.SetTextColor(0, 0, 0)
		pdf.SetFont("", "B", 0)

		for _, program := range programList {
			pdf.SetX(53)
			pdf.CellFormat(100, 10, program.Name, "", 0, "", false, 0, "")
			pdf.CellFormat(150, 10, "version "+program.Version, "", 0, "", false, 0, "")
			pdf.Ln(-1)
		}
		pdf.SetX(80)
		pdf.CellFormat(100, 10, "X Marks a disabled account;", "", 0, "", false, 0, "")

		return err
	}*/

	// Call function
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
	for index, sshConnection := range sshConnectionList {
		list, err := sshConnection.GetWindowsInterfaceInfo()
		fmt.Println(list)
		err = DrawSystemInfoTable(pdf, index, sshConnection)
		if err != nil {
			return err
		}

		err = DrawPhysDriveTable(pdf, index, sshConnection)
		if err != nil {
			log.Println("fail to excute function DrawPhysDriveTable =>", err.Error())
		}

		err = DrawLogicDriveTable(pdf, index, sshConnection)
		if err != nil {
			log.Println("fail to excute function DrawLogicDriveTable =>", err.Error())
		}

		err = DrawLocalUserTable(pdf, index, sshConnection)
		if err != nil {
			log.Println("fail to excute function DrawLocalUserTable =>", err.Error())
		}

		err = DrawWindowsInterfaceTable(pdf, index, sshConnection)
		if err != nil {
			log.Println("fail to excute function DrawWindowsInterfaceTable =>", err.Error())
		}

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
	for index, sshConnection := range sshConnectionList {
		err = DrawSystemInfoTable(pdf, index, sshConnection)
		if err != nil {
			return err
		}

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
	for index, sshConnection := range sshConnectionList {
		err = DrawSystemInfoTable(pdf, index, sshConnection)
		if err != nil {
			return err
		}
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
	for index, sshConnection := range sshConnectionList {
		err = DrawSystemInfoTable(pdf, index, sshConnection)
		if err != nil {
			return err
		}
	}

	return pdf.OutputFileAndClose(filename)
}
