package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/Jeffail/gabs"
	"github.com/jung-kurt/gofpdf"
	"github.com/wcharczuk/go-chart/v2"
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

type ReportModules struct {
	SshConnectionId []int    `json:"sshConnectionId"`
	Modules         []int    `json:"modules"`
	ReceiveEmail    []string `json:"receiveEmail"`
	Cc              []string `json:"cc"`
	Bcc             []string `json:"bcc"`
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

func CountOS() (Report, error) {
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
	return reportInfo, err
}

func ExportReport(filename string, modulesList ReportModules) error {

	// Cover Page
	type recType struct {
		align, txt string
	}

	// Export Pie chart PNG

	report, err := CountOS()
	if err != nil {
		return err
	}
	pie := chart.PieChart{
		Width:  512,
		Height: 512,
		Values: []chart.Value{
			{Value: float64(report.Linux_os_total), Label: "Linux"},
			{Value: float64(report.Netowrk_os_total), Label: "Network"},
			{Value: float64(report.Windows_os_total), Label: "Windows"},
			{Value: float64(report.Unknown_os_total), Label: "Unknown"},
		},
	}

	f, _ := os.Create("./tmp/piechart.png")
	defer f.Close()
	pie.Render(chart.PNG, f)

	var formatRect = func(pdf *gofpdf.Fpdf) {
		pdf.AddPage()
		pdf.SetMargins(10, 10, 10)
		pdf.SetAutoPageBreak(false, 0)
		//borderStr := "1"
		pdf.ImageOptions(
			"./pictures/fpt.png",
			70, 70,
			0, 0,
			false,
			gofpdf.ImageOptions{ImageType: "PNG", ReadDpi: true},
			0,
			"",
		)

		pdf.SetFont("Arial", "B", 22)
		pdf.SetXY(20, 20)
		pdf.CellFormat(258, 365, "Web Application Report", "1", 0, "CM", false, 0, "")
		pdf.SetXY(11, 15)
		pdf.CellFormat(280, 360, "Version 1.0", "", 1, "BC", false, 0, "")
		pdf.SetXY(20, 20)
		pdf.CellFormat(258, 365, utils.GetCurrentDateTime(), "", 0, "BC", false, 0, "")
		// Return Font
		pdf.SetFont("Arial", "", 16)

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
		pdf.SetXY(220, 8)
		pdf.CellFormat(60, 6, "Asset Detail Report", "", 0, "TR", false, 0, "")
		pdf.SetTopMargin(20.0)
		pdf.Ln(10)
	}, true)
	pdf.SetFooterFunc(func() {
		pdf.SetY(-15)
		pdf.SetFont("Arial", "I", 8)
		pdf.CellFormat(0, 10, fmt.Sprintf("Page %d/{nb}", pdf.PageNo()),
			"", 0, "C", false, 0, "")
	})

	pdf.AliasNbPages("")
	pdf.SetFont("Arial", "B", 30)
	formatRect(pdf)

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
		pdf.Ln(8)
	}

	pdf.Ln(2)
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
		pdf.SetY(90)
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
		pdf.SetFont("", "B", 15)
		pdf.WriteAligned(100, 20, "Windows Physical Disk", "L")
		pdf.SetFont("", "", 12)
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
		pdf.SetFont("", "B", 15)
		pdf.WriteAligned(100, 20, "Windows Logical Disk", "L")
		pdf.SetFont("", "", 12)
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

	var DrawWindowsLicenseTable = func(pdf *gofpdf.Fpdf, index int, sshConnection SshConnectionInfo) error {
		pdf.SetAutoPageBreak(true, 20.0)

		// Draw System info
		header := []string{"ProductName", "ProductKey", "ProductId"}

		pdf.Ln(20)
		pdf.SetFont("", "B", 15)
		pdf.WriteAligned(100, 20, "Windows Licenses", "L")
		pdf.SetFont("", "", 12)
		pdf.Ln(20)
		pdf.SetFillColor(141, 151, 173)
		pdf.SetTextColor(255, 255, 255)
		pdf.SetLineWidth(.3)
		pdf.SetFont("", "B", 0)
		w := []float64{70.0, 85.0, 65.0}
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
		licenseList, err := sshConnection.GetAllWindowsLicense()
		if err != nil {
			return err
		}

		for _, license := range licenseList {
			var height float64 = 6
			pdf.CellFormat(w[0], height, license.ProductName, "1", 0, "", fill, 0, "")
			pdf.CellFormat(w[1], height, license.ProductKey, "1", 0, "", fill, 0, "")
			pdf.CellFormat(w[2], height, license.ProductId, "1", 0, "", fill, 0, "")
			pdf.Ln(-1)
			fill = !fill
		}

		return err
	}

	var DrawWindowsLocalUserTable = func(pdf *gofpdf.Fpdf, index int, sshConnection SshConnectionInfo) error {
		pdf.SetAutoPageBreak(true, 20.0)
		pdf.Ln(20)
		pdf.SetFont("", "B", 15)
		pdf.WriteAligned(100, 20, "Windows Local Users", "L")
		pdf.SetFont("", "", 12)
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
		marginCell := 2. // margin of top/bottom of cell
		// Draw System info
		header := []string{"Name", "Description", "IP", "Mac", "DHCPServer", "Subnet", "InterfaceType", "Manufacturer", "DefaultGateway", "DNSDomain"}

		pdf.Ln(20)
		pdf.SetFont("", "B", 15)
		pdf.WriteAligned(100, 20, "Windows Interfaces Information", "L")
		pdf.SetFont("", "", 12)
		pdf.Ln(20)
		pdf.SetFillColor(141, 151, 173)
		pdf.SetTextColor(255, 255, 255)
		pdf.SetLineWidth(.3)
		pdf.SetFont("", "B", 0)
		w := []float64{35.0, 45.0, 32.0, 35.0, 30.0, 35.0, 30.0, 35.0, 35.0, 35.0}
		for j, str := range header {
			if j == 4 || j == 9 {
				continue
			}
			pdf.CellFormat(w[j], 7, str, "1", 0, "C", true, 0, "")
		}
		pdf.Ln(-1)

		// Color and font restoration
		pdf.SetFillColor(224, 235, 255)
		pdf.SetTextColor(0, 0, 0)
		pdf.SetFont("", "", 10)
		// 	Data
		interfacesList, err := sshConnection.GetWindowsInterfaceInfo()
		if err != nil {
			return err
		}

		for _, interfaces := range interfacesList {
			curx, y := pdf.GetXY()
			x := curx

			height := 0.
			_, lineHt := pdf.GetFontSize()

			v := reflect.ValueOf(interfaces)
			typeOfS := v.Type()

			for i := 0; i < typeOfS.NumField(); i++ {
				lines := pdf.SplitLines([]byte(v.Field(i).String()), w[i])
				h := float64(len(lines))*lineHt + marginCell*float64(len(lines))
				if h > height {
					height = h
				}
			}

			for i := 0; i < typeOfS.NumField(); i++ {
				if i == 4 || i == 9 {
					continue
				}
				width := w[i]
				pdf.Rect(x, y, width, height, "")
				pdf.MultiCell(width, lineHt+marginCell, v.Field(i).String(), "", "", false)
				x += width
				pdf.SetXY(x, y)
			}
			pdf.SetXY(curx, y+height)
		}

		return err
	}

	var DrawWindowsDefenderInfoTable = func(pdf *gofpdf.Fpdf, index int, sshConnection SshConnectionInfo) error {
		pdf.SetAutoPageBreak(true, 20.0)
		pdf.Ln(20)
		pdf.SetFont("", "B", 15)
		pdf.WriteAligned(100, 20, "Windows Defender", "L")
		pdf.SetFont("", "", 12)
		pdf.Ln(20)

		// Draw defender info
		pdf.SetFont("Arial", "B", 16)
		defenderInfo, err := sshConnection.GetWindowsDefenderInfo()
		if err != nil {
			return err
		}
		pdf.SetFillColor(141, 151, 173)
		pdf.SetTextColor(255, 255, 255)
		pdf.SetLineWidth(.3)
		pdf.SetFont("", "B", 0)
		pdf.CellFormat(195, 7, "Defender Info", "1", 0, "", true, 0, "")
		pdf.Ln(-1)

		// Color and font restoration
		pdf.SetFillColor(224, 235, 255)
		pdf.SetTextColor(0, 0, 0)
		pdf.SetFont("", "", 0)
		// 	Data
		fill := false
		widthCols := 75.
		heightCols := 6.
		widthContent := 120.
		heightContetnt := 6.
		for i := 0; i < 19; i++ {
			pdf.SetFont("", "B", 12)
			switch i {
			case 0:
				pdf.CellFormat(widthCols, heightCols, "AMEngineVersion", "1", 0, "", fill, 0, "")
				pdf.SetFont("", "", 12)
				pdf.CellFormat(widthContent, heightContetnt, defenderInfo.AMEngineVersion, "1", 0, "", fill, 0, "")
				pdf.Ln(-1)
			case 1:
				pdf.CellFormat(widthCols, heightCols, "AMProductVersion", "1", 0, "", fill, 0, "")
				pdf.SetFont("", "", 12)
				pdf.CellFormat(widthContent, heightContetnt, defenderInfo.AMProductVersion, "1", 0, "", fill, 0, "")
				pdf.Ln(-1)
			case 2:
				pdf.CellFormat(widthCols, heightCols, "AMServiceEnabled", "1", 0, "", fill, 0, "")
				pdf.SetFont("", "", 12)
				pdf.CellFormat(widthContent, heightContetnt, strconv.FormatBool(defenderInfo.AMServiceEnabled), "1", 0, "", fill, 0, "")
				pdf.Ln(-1)
			case 3:
				pdf.CellFormat(widthCols, heightCols, "AntispywareEnabled", "1", 0, "", fill, 0, "")
				pdf.SetFont("", "", 12)
				pdf.CellFormat(widthContent, heightContetnt, strconv.FormatBool(defenderInfo.AntispywareEnabled), "1", 0, "", fill, 0, "")
				pdf.Ln(-1)
			case 4:
				pdf.CellFormat(widthCols, heightCols, "AntispywareSignatureLastUpdated", "1", 0, "", fill, 0, "")
				pdf.SetFont("", "", 12)
				pdf.CellFormat(widthContent, heightContetnt, defenderInfo.AntispywareSignatureLastUpdated, "1", 0, "", fill, 0, "")
				pdf.Ln(-1)
			case 5:
				pdf.CellFormat(widthCols, heightCols, "AntispywareSignatureVersion", "1", 0, "", fill, 0, "")
				pdf.SetFont("", "", 12)
				pdf.CellFormat(widthContent, heightContetnt, defenderInfo.AntispywareSignatureVersion, "1", 0, "", fill, 0, "")
				pdf.Ln(-1)
			case 6:
				pdf.CellFormat(widthCols, heightCols, "AntivirusEnabled", "1", 0, "", fill, 0, "")
				pdf.SetFont("", "", 12)
				pdf.CellFormat(widthContent, heightContetnt, strconv.FormatBool(defenderInfo.AntivirusEnabled), "1", 0, "", fill, 0, "")
				pdf.Ln(-1)
			case 7:
				pdf.CellFormat(widthCols, heightCols, "AntivirusSignatureLastUpdated", "1", 0, "", fill, 0, "")
				pdf.SetFont("", "", 12)
				pdf.CellFormat(widthContent, heightContetnt, defenderInfo.AntivirusSignatureLastUpdated, "1", 0, "", fill, 0, "")
				pdf.Ln(-1)
			case 8:
				pdf.CellFormat(widthCols, heightCols, "AntivirusSignatureVersion", "1", 0, "", fill, 0, "")
				pdf.SetFont("", "", 12)
				pdf.CellFormat(widthContent, heightContetnt, defenderInfo.AntivirusSignatureVersion, "1", 0, "", fill, 0, "")
				pdf.Ln(-1)
			case 9:
				pdf.CellFormat(widthCols, heightCols, "BehaviorMonitorEnabled", "1", 0, "", fill, 0, "")
				pdf.SetFont("", "", 12)
				pdf.CellFormat(widthContent, heightContetnt, strconv.FormatBool(defenderInfo.BehaviorMonitorEnabled), "1", 0, "", fill, 0, "")
				pdf.Ln(-1)
			case 10:
				pdf.CellFormat(widthCols, heightCols, "ComputerState", "1", 0, "", fill, 0, "")
				pdf.SetFont("", "", 12)
				pdf.CellFormat(widthContent, heightContetnt, defenderInfo.ComputerState, "1", 0, "", fill, 0, "")
				pdf.Ln(-1)
			case 11:
				pdf.CellFormat(widthCols, heightCols, "FullScanAge", "1", 0, "", fill, 0, "")
				pdf.SetFont("", "", 12)
				pdf.CellFormat(widthContent, heightContetnt, defenderInfo.FullScanAge, "1", 0, "", fill, 0, "")
				pdf.Ln(-1)
			case 12:
				pdf.CellFormat(widthCols, heightCols, "IoavProtectionEnabled", "1", 0, "", fill, 0, "")
				pdf.SetFont("", "", 12)
				pdf.CellFormat(widthContent, heightContetnt, strconv.FormatBool(defenderInfo.IoavProtectionEnabled), "1", 0, "", fill, 0, "")
				pdf.Ln(-1)
			case 13:
				pdf.CellFormat(widthCols, heightCols, "IsTamperProtected", "1", 0, "", fill, 0, "")
				pdf.SetFont("", "", 12)
				pdf.CellFormat(widthContent, heightContetnt, strconv.FormatBool(defenderInfo.IsTamperProtected), "1", 0, "", fill, 0, "")
				pdf.Ln(-1)
			case 14:
				pdf.CellFormat(widthCols, heightCols, "NISEnabled", "1", 0, "", fill, 0, "")
				pdf.SetFont("", "", 12)
				pdf.CellFormat(widthContent, heightContetnt, strconv.FormatBool(defenderInfo.NISEnabled), "1", 0, "", fill, 0, "")
				pdf.Ln(-1)
			case 15:
				pdf.CellFormat(widthCols, heightCols, "NISEngineVersion", "1", 0, "", fill, 0, "")
				pdf.SetFont("", "", 12)
				pdf.CellFormat(widthContent, heightContetnt, defenderInfo.NISEngineVersion, "1", 0, "", fill, 0, "")
				pdf.Ln(-1)
			case 16:
				pdf.CellFormat(widthCols, heightCols, "NISSignatureLastUpdated", "1", 0, "", fill, 0, "")
				pdf.SetFont("", "", 12)
				pdf.CellFormat(widthContent, heightContetnt, defenderInfo.NISSignatureLastUpdated, "1", 0, "", fill, 0, "")
				pdf.Ln(-1)
			case 17:
				pdf.CellFormat(widthCols, heightCols, "OnAccessProtectionEnabled", "1", 0, "", fill, 0, "")
				pdf.SetFont("", "", 12)
				pdf.CellFormat(widthContent, heightContetnt, strconv.FormatBool(defenderInfo.OnAccessProtectionEnabled), "1", 0, "", fill, 0, "")
				pdf.Ln(-1)
			case 18:
				pdf.CellFormat(widthCols, heightCols, "LastQuickScan", "1", 0, "", fill, 0, "")
				pdf.SetFont("", "", 12)
				pdf.CellFormat(widthContent, heightContetnt, defenderInfo.LastQuickScan, "1", 0, "", fill, 0, "")
				pdf.Ln(-1)
			case 19:
				pdf.CellFormat(widthCols, heightCols, "RealTimeProtectionEnabled", "1", 0, "", fill, 0, "")
				pdf.SetFont("", "", 12)
				pdf.CellFormat(widthContent, heightContetnt, strconv.FormatBool(defenderInfo.RealTimeProtectionEnabled), "1", 0, "", fill, 0, "")
				pdf.Ln(-1)

			}
			fill = !fill
		}
		return err
	}

	var DrawLinuxInterfaceTable = func(pdf *gofpdf.Fpdf, index int, sshConnection SshConnectionInfo) error {
		pdf.SetAutoPageBreak(true, 20.0)
		marginCell := 2. // margin of top/bottom of cell
		// Draw System info
		header := []string{"Active", "InterfaceName", "Address", "Broadcast", "Netmask", "Network", "Gateway", "Interface", "Mac", "InterfaceType"}

		pdf.Ln(20)
		pdf.SetFont("", "B", 15)
		pdf.WriteAligned(100, 20, "Linux Interfaces Information", "L")
		pdf.SetFont("", "", 12)
		pdf.Ln(20)
		pdf.SetFillColor(141, 151, 173)
		pdf.SetTextColor(255, 255, 255)
		pdf.SetLineWidth(.3)
		pdf.SetFont("", "B", 0)
		w := []float64{20.0, 30.0, 30.0, 30.0, 30.0, 30.0, 25.0, 25.0, 32.0, 30.0}
		for j, str := range header {
			pdf.CellFormat(w[j], 7, str, "1", 0, "C", true, 0, "")
		}
		pdf.Ln(-1)

		// Color and font restoration
		pdf.SetFillColor(224, 235, 255)
		pdf.SetTextColor(0, 0, 0)
		pdf.SetFont("", "", 10)
		// 	Data
		interfacesList, err := sshConnection.GetLinuxInterfaceInfo()
		if err != nil {
			return err
		}
		type tmpInterface struct {
			Active        string `json:"active"`
			InterfaceName string `json:"device"`
			Address       string `json:"address"`
			Broadcast     string `json:"broadcast"`
			Netmask       string `json:"netmask"`
			Network       string `json:"network"`
			DefautGateway string `json:"gateway"`
			Interface     string `json:"interface"`
			Mac           string `json:"macaddress"`
			InterfaceType string `json:"type"`
		}
		var tmpInterfaceList []tmpInterface
		for _, interfaces := range interfacesList {
			var tmpInterface tmpInterface
			tmpInterface.Active = strconv.FormatBool(interfaces.Active)
			tmpInterface.InterfaceName = interfaces.InterfaceName
			tmpInterface.Address = interfaces.IPv4.Address
			tmpInterface.Broadcast = interfaces.IPv4.Broadcast
			tmpInterface.Netmask = interfaces.IPv4.Netmask
			tmpInterface.Network = interfaces.IPv4.Network
			tmpInterface.DefautGateway = interfaces.IPv4.DefautGateway
			tmpInterface.Interface = interfaces.IPv4.Interface
			tmpInterface.Mac = interfaces.Mac
			tmpInterface.InterfaceType = interfaces.InterfaceType
			tmpInterfaceList = append(tmpInterfaceList, tmpInterface)
		}

		_, pageh := pdf.GetPageSize()
		_, _, _, mbottom := pdf.GetMargins()

		for _, interfaces := range tmpInterfaceList {
			curx, y := pdf.GetXY()
			x := curx

			height := 0.
			_, lineHt := pdf.GetFontSize()

			v := reflect.ValueOf(interfaces)
			typeOfS := v.Type()

			for i := 0; i < typeOfS.NumField(); i++ {
				lines := pdf.SplitLines([]byte(v.Field(i).String()), w[i])
				h := float64(len(lines))*lineHt + marginCell*float64(len(lines))
				if h > height {
					height = h
				}
			}

			if pdf.GetY()+height > pageh-mbottom {
				pdf.AddPage()
				pdf.SetY(30)
				y = pdf.GetY()
			}

			for i := 0; i < typeOfS.NumField(); i++ {
				width := w[i]
				pdf.Rect(x, y, width, height, "")
				pdf.MultiCell(width, lineHt+marginCell, v.Field(i).String(), "", "", false)
				x += width
				pdf.SetXY(x, y)
			}
			pdf.SetXY(curx, y+height)
		}

		return err
	}

	var DrawWindowsProgramTable = func(pdf *gofpdf.Fpdf, index int, sshConnection SshConnectionInfo) error {
		pdf.SetAutoPageBreak(true, 20.0)
		pdf.Ln(20)
		pdf.SetFont("", "B", 15)
		pdf.WriteAligned(100, 20, "Windows Program", "L")
		pdf.SetFont("", "", 12)
		pdf.Ln(20)
		programList, err := sshConnection.GetInstalledProgram()
		if err != nil {
			return err
		}

		pdf.SetFillColor(141, 151, 173)
		pdf.SetTextColor(255, 255, 255)
		pdf.SetLineWidth(.3)
		pdf.SetFont("", "B", 0)
		pdf.CellFormat(280, 10, "Software Versions and Usage", "1", 0, "C", true, 0, "")
		pdf.Ln(-1)
		// Color and font restoration
		pdf.SetFillColor(224, 235, 255)
		pdf.SetTextColor(0, 0, 0)
		pdf.SetFont("", "", 9)

		for i := 0; i < len(programList); i++ {
			pdf.CellFormat(100, 10, programList[i].Name, "", 0, "L", false, 0, "")
			pdf.CellFormat(40, 10, "version "+programList[i].Version, "", 0, "L", false, 0, "")

			if i+1 <= len(programList)-1 {
				pdf.CellFormat(100, 10, "| "+programList[i].Name, "", 0, "L", false, 0, "")
				pdf.CellFormat(30, 10, "version "+programList[i].Version, "", 0, "L", false, 0, "")
			}
			pdf.Ln(-1)
		}

		return err
	}

	var DrawLinuxLocalUserTable = func(pdf *gofpdf.Fpdf, index int, sshConnection SshConnectionInfo) error {
		pdf.SetAutoPageBreak(true, 20.0)
		pdf.Ln(20)
		pdf.SetFont("", "B", 15)
		pdf.WriteAligned(100, 20, "Linux Local Users", "L")
		pdf.SetFont("", "", 12)
		pdf.Ln(20)
		localUserList, err := sshConnection.GetLinuxUsersLastLogin()
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
			pdf.SetX(62)
			pdf.CellFormat(100, 10, localUser.Username, "", 0, "", false, 0, "")
			pdf.CellFormat(150, 10, localUser.LastLoginTime, "", 0, "", false, 0, "")
			pdf.Ln(-1)
		}

		return err
	}

	var DrawNetworkIpAddressTable = func(pdf *gofpdf.Fpdf, index int, sshConnection SshConnectionInfo) error {
		pdf.SetAutoPageBreak(true, 20.0)

		// Draw System info
		header := []string{"Index", "IpInterface", "Address", "NetMask", "BcastAddr", "ReasmMaxSize"}

		pdf.Ln(20)
		pdf.SetFont("", "B", 15)
		pdf.WriteAligned(100, 20, "Network Address", "L")
		pdf.SetFont("", "", 12)
		pdf.Ln(20)
		pdf.SetFillColor(141, 151, 173)
		pdf.SetTextColor(255, 255, 255)
		pdf.SetLineWidth(.3)
		pdf.SetFont("", "B", 0)
		w := []float64{15.0, 124.0, 35.0, 40.0, 30.0, 32.0}
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
		ipAddrList, err := GetNetworkIPAddr(sshConnection.SSHConnectionId)
		if err != nil {
			return err
		}

		for _, ipAddr := range ipAddrList {
			var height float64 = 6
			pdf.CellFormat(w[0], height, strconv.Itoa(ipAddr.IpAdEntIfIndex), "1", 0, "", fill, 0, "")
			pdf.CellFormat(w[1], height, ipAddr.IpInterface, "1", 0, "", fill, 0, "")
			pdf.CellFormat(w[2], height, ipAddr.IpAdEntAddr, "1", 0, "", fill, 0, "")
			pdf.CellFormat(w[3], height, ipAddr.IpAdEntNetMask, "1", 0, "", fill, 0, "")
			pdf.CellFormat(w[4], height, strconv.Itoa(ipAddr.IpAdEntBcastAddr), "1", 0, "", fill, 0, "")
			pdf.CellFormat(w[5], height, strconv.Itoa(ipAddr.IpAdEntReasmMaxSize), "1", 0, "", fill, 0, "")
			pdf.Ln(-1)
			fill = !fill
		}

		return err
	}

	var DrawEndPage = func(pdf *gofpdf.Fpdf) error {
		pdf.SetAutoPageBreak(true, 20.0)
		pdf.AddPage()
		pdf.SetMargins(10, 10, 10)
		pdf.SetAutoPageBreak(false, 0)
		pdf.ImageOptions(
			"./pictures/fpt.png",
			70, 70,
			0, 0,
			false,
			gofpdf.ImageOptions{ImageType: "PNG", ReadDpi: true},
			0,
			"",
		)

		pdf.SetFont("Arial", "B", 22)
		pdf.SetXY(20, 20)
		pdf.CellFormat(258, 365, "End Report", "1", 0, "CM", false, 0, "")
		pdf.SetXY(20, 20)
		pdf.CellFormat(258, 365, utils.GetCurrentDateTime(), "", 0, "BC", false, 0, "")

		return err
	}

	var DrawOSSummaryTable = func(pdf *gofpdf.Fpdf, osType string) error {
		pdf.SetAutoPageBreak(true, 20.0)

		// Draw System info
		header := []string{"SSHConnectionId", "OsName", "Hostname", "OsVersion"}

		pdf.Ln(20)
		pdf.SetFont("", "B", 15)
		pdf.WriteAligned(100, 20, osType+" OS Summary", "L")
		pdf.SetFont("", "", 12)
		pdf.Ln(20)
		pdf.SetFillColor(141, 151, 173)
		pdf.SetTextColor(255, 255, 255)
		pdf.SetLineWidth(.3)
		pdf.SetFont("", "B", 0)
		w := []float64{40.0, 60.0, 72.0, 75.0}
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
		detailInfoList, err := GetDetailOSReport(osType)
		if err != nil {
			return err
		}

		for _, detailInfo := range detailInfoList {
			var height float64 = 6
			pdf.CellFormat(w[0], height, strconv.Itoa(detailInfo.SshConnectionId), "1", 0, "", fill, 0, "")
			pdf.CellFormat(w[1], height, detailInfo.OsName, "1", 0, "", fill, 0, "")
			pdf.CellFormat(w[2], height, detailInfo.Hostname, "1", 0, "", fill, 0, "")
			pdf.CellFormat(w[3], height, detailInfo.OsVersion, "1", 0, "", fill, 0, "")
			pdf.Ln(-1)
			fill = !fill
		}

		return err
	}

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

	modules := modulesList.Modules
	// Get Windows
	sshConnectionList, err = GetAllOSSSHConnection("Windows")
	if err != nil {
		return err
	}
	for index, sshConnection := range sshConnectionList {
		if CheckModules(modules, 1) {
			err = DrawSystemInfoTable(pdf, index, sshConnection)
			if err != nil {
				log.Println(sshConnection.HostNameSSH, ":fail to excute function DrawSystemInfoTable =>", err.Error())
			}
		}

		if CheckModules(modules, 2) {
			err = DrawPhysDriveTable(pdf, index, sshConnection)
			if err != nil {
				log.Println(sshConnection.HostNameSSH, ":fail to excute function DrawPhysDriveTable =>", err.Error())
			}
		}

		if CheckModules(modules, 3) {
			err = DrawLogicDriveTable(pdf, index, sshConnection)
			if err != nil {
				log.Println(sshConnection.HostNameSSH, ":fail to excute function DrawLogicDriveTable =>", err.Error())
			}
		}

		if CheckModules(modules, 4) {
			err = DrawWindowsLocalUserTable(pdf, index, sshConnection)
			if err != nil {
				log.Println(sshConnection.HostNameSSH, ":fail to excute function DrawLocalUserTable =>", err.Error())
			}
		}

		if CheckModules(modules, 5) {
			err = DrawWindowsInterfaceTable(pdf, index, sshConnection)
			if err != nil {
				log.Println(sshConnection.HostNameSSH, ":fail to excute function DrawWindowsInterfaceTable =>", err.Error())
			}
		}

		if CheckModules(modules, 6) {
			err = DrawWindowsProgramTable(pdf, index, sshConnection)
			if err != nil {
				log.Println(sshConnection.HostNameSSH, ":fail to excute function DrawWindowsProgramTable =>", err.Error())
			}
		}

		if CheckModules(modules, 7) {
			err = DrawWindowsDefenderInfoTable(pdf, index, sshConnection)
			if err != nil {
				log.Println(sshConnection.HostNameSSH, ":fail to excute function DrawWindowsDefenderInfoTable =>", err.Error())
			}
		}

		if CheckModules(modules, 8) {
			err = DrawWindowsLicenseTable(pdf, index, sshConnection)
			if err != nil {
				log.Println(sshConnection.HostNameSSH, ":fail to excute function DrawWindowsLicenseTable =>", err.Error())
			}
		}

	}
	if CheckModules(modules, 9) {
		err = DrawOSSummaryTable(pdf, "Windows")
		if err != nil {
			log.Println("fail to excute function DrawOSSummaryTable =>", err.Error())
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

		if CheckModules(modules, 10) {
			err = DrawSystemInfoTable(pdf, index, sshConnection)
			if err != nil {
				log.Println(sshConnection.HostNameSSH, ":fail to excute function DrawSystemInfoTable =>", err.Error())
			}
		}

		if CheckModules(modules, 11) {
			err = DrawLinuxLocalUserTable(pdf, index, sshConnection)
			if err != nil {
				log.Println(sshConnection.HostNameSSH, ":fail to excute function DrawLocalUserTable =>", err.Error())
			}
		}

		if CheckModules(modules, 12) {
			err = DrawLinuxInterfaceTable(pdf, index, sshConnection)
			if err != nil {
				log.Println(sshConnection.HostNameSSH, ":fail to excute function DrawWindowsInterfaceTable =>", err.Error())
			}
		}

	}
	if CheckModules(modules, 13) {
		err = DrawOSSummaryTable(pdf, "Linux")
		if err != nil {
			log.Println("fail to excute function DrawOSSummaryTable =>", err.Error())
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

		if CheckModules(modules, 20) {
			err = DrawSystemInfoTable(pdf, index, sshConnection)
			if err != nil {
				log.Println(sshConnection.HostNameSSH, ":fail to excute function DrawSystemInfoTable =>", err.Error())
			}
		}

		if CheckModules(modules, 21) {
			err = DrawNetworkIpAddressTable(pdf, index, sshConnection)
			if err != nil {
				log.Println(sshConnection.HostNameSSH, ":fail to excute function DrawNetworkIpAddressTable =>", err.Error())
			}
		}
	}
	if CheckModules(modules, 22) {
		err = DrawOSSummaryTable(pdf, "Router")
		if err != nil {
			log.Println("fail to excute function DrawOSSummaryTable =>", err.Error())
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

		if CheckModules(modules, 30) {
			err = DrawSystemInfoTable(pdf, index, sshConnection)
			if err != nil {
				log.Println(sshConnection.HostNameSSH, ":fail to excute function DrawSystemInfoTable =>", err.Error())
			}
		}

		if CheckModules(modules, 31) {
			err = DrawNetworkIpAddressTable(pdf, index, sshConnection)
			if err != nil {
				log.Println(sshConnection.HostNameSSH, ":fail to excute function DrawNetworkIpAddressTable =>", err.Error())
			}
		}
	}
	if CheckModules(modules, 32) {
		err = DrawOSSummaryTable(pdf, "Switch")
		if err != nil {
			log.Println("fail to excute function DrawOSSummaryTable =>", err.Error())
		}
	}

	// End Page
	err = DrawEndPage(pdf)
	if err != nil {
		log.Println("fail to excute function DrawEndPage =>", err.Error())
	}
	return pdf.OutputFileAndClose(filename)
}

func CheckModules(modules []int, number int) bool {
	for _, module := range modules {
		if module == number {
			return true
		}
	}
	return false
}
