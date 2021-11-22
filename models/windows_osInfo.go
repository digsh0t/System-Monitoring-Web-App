package models

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/Jeffail/gabs"
)

type osProfile struct {
	OS            string         `json:"os"`
	OSKey         windowsLicense `json:"os_key"`
	Manufacturer  string         `json:"manufacturer"`
	Model         string         `json:"model"`
	SerialNumber  string         `json:"serial_number"`
	Processor     string         `json:"processor"`
	OSInstallDate string         `json:"os_install_date"`
}

type physicalDrive struct {
	Name         string `json:"name"`
	Serial       string `json:"serial"`
	Manufacturer string `json:"manufacturer"`
	Model        string `json:"hardware_model"`
	Description  string `json:"description"`
	Type         string `json:"type"`
	Partition    string `json:"partitions"`
	DiskSize     string `json:"disk_size"`
}

type logicalDrive struct {
	Name        string `json:"device_id"`
	Description string `json:"description"`
	Type        string `json:"type"`
	FileSystem  string `json:"file_system"`
	Size        string `json:"size"`
	FreeSpace   string `json:"free_space"`
}

type windowsDefenderStatus struct {
	AMEngineVersion                 string
	AMProductVersion                string
	AMServiceEnabled                bool
	AntispywareEnabled              bool
	AntispywareSignatureLastUpdated string
	AntispywareSignatureVersion     string
	AntivirusEnabled                bool
	AntivirusSignatureLastUpdated   string
	AntivirusSignatureVersion       string
	BehaviorMonitorEnabled          bool
	ComputerState                   string
	FullScanAge                     string
	IoavProtectionEnabled           bool
	IsTamperProtected               bool
	NISEnabled                      bool
	NISEngineVersion                string
	NISSignatureLastUpdated         string
	OnAccessProtectionEnabled       bool
	LastQuickScan                   string
	RealTimeProtectionEnabled       bool
}

type ansibleWindowsInterfaceInfo struct {
	ConnectionName string `json:"connection_name"`
	Description    string `json:"description"`
	IP             string `json:"ipv4_address"`
	Mac            string `json:"mac"`
	DHCPServer     string `json:"dhcp_server"`
	Subnet         string `json:"mask"`
	InterfaceType  string `json:"type"`
	Manufacturer   string `json:"manufacturer"`
	DefaultGateway string `json:"default_gateway"`
	DNSDomain      string `json:"dns_domain"`
}

type ansibleLinuxInterfaceInfo struct {
	Active        bool        `json:"active"`
	InterfaceName string      `json:"device"`
	IPv4          addressV4   `json:"ipv4"`
	IPV6          []addressV6 `json:"ipv6"`
	Mac           string      `json:"macaddress"`
	InterfaceType string      `json:"type"`
}

type addressV4 struct {
	Address       string `json:"address"`
	Broadcast     string `json:"broadcast"`
	Netmask       string `json:"netmask"`
	Network       string `json:"network"`
	DefautGateway string `json:"gateway"`
	Interface     string `json:"interface"`
}

type addressV6 struct {
	Address string `json:"address"`
	Prefix  string `json:"prefix"`
	Scope   string `json:"scope"`
}

var InterfaceTypes = map[string]string{
	"6":  "ethernet-csmacd",
	"24": "softwareLoopback",
}

func (sshConnection SshConnectionInfo) GetWindowsPhysicalDiskInfo() ([]physicalDrive, error) {
	var driveList []physicalDrive
	result, err := sshConnection.RunCommandFromSSHConnectionUseKeys(`osqueryi --json "SELECT * FROM disk_info";`)
	if err != nil {
		return driveList, err
	}
	err = json.Unmarshal([]byte(result), &driveList)
	if err != nil {
		return driveList, err
	}
	return driveList, err
}

func (sshConnection SshConnectionInfo) GetWindowsLogicalDriveInfo() ([]logicalDrive, error) {
	var logicalDriveList []logicalDrive
	result, err := sshConnection.RunCommandFromSSHConnectionUseKeys(`osqueryi --json "SELECT * FROM logical_drives";`)
	if err != nil {
		return logicalDriveList, err
	}
	err = json.Unmarshal([]byte(result), &logicalDriveList)
	if err != nil {
		return logicalDriveList, err
	}
	return logicalDriveList, err
}

func (sshConnection SshConnectionInfo) GetWindowsDefenderInfo() (windowsDefenderStatus, error) {
	var defenderStatus windowsDefenderStatus
	result, err := sshConnection.RunCommandFromSSHConnectionUseKeys(`powershell -Command "Get-MpComputerStatus | select AMEngineVersion,AMProductVersion,AMServiceEnabled,AntispywareEnabled,AntispywareSignatureLastUpdated,AntispywareSignatureVersion,AntivirusEnabled,AntivirusSignatureLastUpdated,AntivirusSignatureVersion,BehaviorMonitorEnabled,ComputerState,FullScanAge,IoavProtectionEnabled,IsTamperProtected,NISEnabled,NISEngineVersion,NISSignatureLastUpdated,OnAccessProtectionEnabled,QuickScanStartTime,RealTimeProtectionEnabled"`)
	if err != nil {
		return defenderStatus, err
	}
	defenderStatus = parseWindowsDefenderInfoOutput(result)
	return defenderStatus, err
}

func parseWindowsDefenderInfoOutput(input string) windowsDefenderStatus {
	var defenderStatus windowsDefenderStatus
	lines := strings.Split(strings.Trim(input, "\r\n\t"), "\n")
	for i := 0; i < len(lines); i++ {
		lines[i] = strings.Trim(strings.Split(lines[i], ":")[1], "\r\n\t ")
	}
	defenderStatus.AMEngineVersion = lines[0]
	defenderStatus.AMProductVersion = lines[1]
	if lines[2] == "True" {
		defenderStatus.AMServiceEnabled = true
	}
	if lines[3] == "True" {
		defenderStatus.AntispywareEnabled = true
	}
	defenderStatus.AntispywareSignatureLastUpdated = lines[4]
	defenderStatus.AntispywareSignatureVersion = lines[5]
	if lines[6] == "True" {
		defenderStatus.AntivirusEnabled = true
	}
	defenderStatus.AntivirusSignatureLastUpdated = lines[7]
	defenderStatus.AntivirusSignatureVersion = lines[8]
	if lines[9] == "True" {
		defenderStatus.BehaviorMonitorEnabled = true
	}
	switch lines[10] {
	case "0":
		defenderStatus.ComputerState = "clean"
	case "1":
		defenderStatus.ComputerState = "pending full scan"
	case "2":
		defenderStatus.ComputerState = "pending reboot"
	case "4":
		defenderStatus.ComputerState = "pending manual steps"
	case "8":
		defenderStatus.ComputerState = "pending offline scan"
	case "16":
		defenderStatus.ComputerState = "pending critical failure"
	}
	defenderStatus.FullScanAge = lines[11]
	if lines[12] == "True" {
		defenderStatus.IoavProtectionEnabled = true
	}
	if lines[13] == "True" {
		defenderStatus.IsTamperProtected = true
	}
	if lines[14] == "True" {
		defenderStatus.NISEnabled = true
	}
	defenderStatus.NISEngineVersion = lines[15]
	defenderStatus.NISSignatureLastUpdated = lines[16]
	if lines[17] == "True" {
		defenderStatus.OnAccessProtectionEnabled = true
	}
	defenderStatus.LastQuickScan = lines[18]
	if lines[19] == "True" {
		defenderStatus.RealTimeProtectionEnabled = true
	}
	return defenderStatus
}

func (sshConnection SshConnectionInfo) GetWindowsInterfaceInfo() ([]ansibleWindowsInterfaceInfo, error) {
	var tmpList []ansibleWindowsInterfaceInfo
	interfaceList, err := sshConnection.getWindowsInterfaceIPInfo()
	if err != nil {
		return nil, err
	}
	output, err := sshConnection.RunAnsiblePlaybookWithjson("./yamls/get_interface_info.yml", `{"host":"`+sshConnection.HostNameSSH+`"}`)
	if err != nil {
		return nil, err
	}
	tmpList, err = parseAnsibleWindowsInterfaceInfoOutput(output)
	for i := 0; i < len(interfaceList); i++ {
		for _, tmp := range tmpList {
			if tmp.ConnectionName == interfaceList[i].ConnectionName {
				interfaceList[i].DNSDomain = tmp.DNSDomain
				interfaceList[i].DefaultGateway = tmp.DefaultGateway
			}
		}
		interfaceList[i].InterfaceType = InterfaceTypes[interfaceList[i].InterfaceType]
	}
	return interfaceList, err
}

func (sshConnection SshConnectionInfo) GetLinuxInterfaceInfo() ([]ansibleLinuxInterfaceInfo, error) {
	var interfaceList []ansibleLinuxInterfaceInfo
	output, err := sshConnection.RunAnsiblePlaybookWithjson("./yamls/get_interface_info.yml", `{"host":"`+sshConnection.HostNameSSH+`"}`)
	if err != nil {
		return nil, err
	}
	interfaceList, err = parseAnsibleLinuxInterfaceInfoOutput(output)
	return interfaceList, err
}

func parseAnsibleWindowsInterfaceInfoOutput(input string) ([]ansibleWindowsInterfaceInfo, error) {

	var interfaceList []ansibleWindowsInterfaceInfo

	re, err := regexp.Compile(`\{[\s\S]*\}`)
	if err != nil {
		return nil, err
	}

	input = re.FindString(input)
	jsonParsed, err := gabs.ParseJSON([]byte(input))
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal([]byte(jsonParsed.Path("ansible_facts.interfaces").String()), &interfaceList)
	return interfaceList, err
}

func (sshConnection SshConnectionInfo) getWindowsInterfaceIPInfo() ([]ansibleWindowsInterfaceInfo, error) {

	var interfaceList []ansibleWindowsInterfaceInfo

	output, err := sshConnection.RunCommandFromSSHConnectionUseKeys(`osqueryi --json "SELECT IA.friendly_name AS 'connection_name',IA.address AS 'ipv4_address',ID.mac,ID.dhcp_server,IA.mask,ID.type,ID.manufacturer,ID.description FROM interface_addresses AS IA LEFT JOIN interface_details AS ID ON IA.interface = ID.interface  WHERE IA.address NOT LIKE '%::%'`)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal([]byte(output), &interfaceList)
	return interfaceList, err
}

func parseAnsibleLinuxInterfaceInfoOutput(input string) ([]ansibleLinuxInterfaceInfo, error) {

	var interfaceList []ansibleLinuxInterfaceInfo
	var interfaceNameList []string
	var tmpInterface ansibleLinuxInterfaceInfo
	var tmpIPv4 addressV4
	var jsonPath string

	re, err := regexp.Compile(`\{[\s\S]*\}`)
	if err != nil {
		return nil, err
	}

	input = re.FindString(input)
	jsonParsed, err := gabs.ParseJSON([]byte(input))
	if err != nil {
		return nil, err
	}
	tmpGaps, _ := jsonParsed.Search("ansible_facts", "interfaces").Children()
	for _, child := range tmpGaps {
		interfaceNameList = append(interfaceNameList, strings.ReplaceAll(child.String(), `"`, ""))
	}
	for _, name := range interfaceNameList {
		jsonPath = "ansible_facts." + name
		err = json.Unmarshal([]byte(jsonParsed.Path(jsonPath).String()), &tmpInterface)
		if err != nil {
			return nil, err
		}
		interfaceList = append(interfaceList, tmpInterface)
		tmpInterface = ansibleLinuxInterfaceInfo{}
	}
	err = json.Unmarshal([]byte(jsonParsed.Path("ansible_facts.default_ipv4").String()), &tmpIPv4)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(interfaceList); i++ {
		if interfaceList[i].InterfaceName == tmpIPv4.Interface {
			interfaceList[i].IPv4 = tmpIPv4
		}
	}
	return interfaceList, err
}
