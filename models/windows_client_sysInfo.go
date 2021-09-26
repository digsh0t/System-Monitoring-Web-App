package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/Jeffail/gabs"
)

type osVersion struct {
	Arch         string `json:"arch"`
	Build        string `json:"build"`
	Codename     string `json:"codename"`
	InstallDate  string `json:"install_date"`
	Major        string `json:"major"`
	Minor        string `json:"minor"`
	Name         string `json:"name"`
	Patch        string `json:"patch"`
	Platform     string `json:"platform"`
	PlatformLike string `json:"platform_like"`
	Version      string `json:"version"`
}

type interfaceInfo struct {
	Collisions                 string `json:"collisions"`
	ConnectionID               string `json:"connection_id"`
	ConnectionStatus           string `json:"connection_status"`
	Description                string `json:"description"`
	DhcpEnabled                string `json:"dhcp_enabled"`
	DhcpLeaseExpires           string `json:"dhcp_lease_expires"`
	DhcpLeaseObtained          string `json:"dhcp_lease_obtained"`
	DhcpServer                 string `json:"dhcp_server"`
	DNSDomain                  string `json:"dns_domain"`
	DNSDomainSuffixSearchOrder string `json:"dns_domain_suffix_search_order"`
	DNSHostName                string `json:"dns_host_name"`
	DNSServerSearchOrder       string `json:"dns_server_search_order"`
	Enabled                    string `json:"enabled"`
	Flags                      string `json:"flags"`
	FriendlyName               string `json:"friendly_name"`
	Ibytes                     string `json:"ibytes"`
	Idrops                     string `json:"idrops"`
	Ierrors                    string `json:"ierrors"`
	Interface                  string `json:"interface"`
	Ipackets                   string `json:"ipackets"`
	LastChange                 string `json:"last_change"`
	Mac                        string `json:"mac"`
	Manufacturer               string `json:"manufacturer"`
	Metric                     string `json:"metric"`
	Mtu                        string `json:"mtu"`
	Obytes                     string `json:"obytes"`
	Odrops                     string `json:"odrops"`
	Oerrors                    string `json:"oerrors"`
	Opackets                   string `json:"opackets"`
	PhysicalAdapter            string `json:"physical_adapter"`
	Service                    string `json:"service"`
	Speed                      string `json:"speed"`
	Type                       string `json:"type"`
}
type cpuInfo struct {
	AddressWidth      string `json:"address_width"`
	Availability      string `json:"availability"`
	CPUStatus         string `json:"cpu_status"`
	CurrentClockSpeed string `json:"current_clock_speed"`
	DeviceID          string `json:"device_id"`
	LogicalProcessors string `json:"logical_processors"`
	Manufacturer      string `json:"manufacturer"`
	MaxClockSpeed     string `json:"max_clock_speed"`
	Model             string `json:"model"`
	NumberOfCores     string `json:"number_of_cores"`
	ProcessorType     string `json:"processor_type"`
	SocketDesignation string `json:"socket_designation"`
}
type connectivity struct {
	Disconnected     string `json:"disconnected"`
	Ipv4Internet     string `json:"ipv4_internet"`
	Ipv4LocalNetwork string `json:"ipv4_local_network"`
	Ipv4NoTraffic    string `json:"ipv4_no_traffic"`
	Ipv4Subnet       string `json:"ipv4_subnet"`
	Ipv6Internet     string `json:"ipv6_internet"`
	Ipv6LocalNetwork string `json:"ipv6_local_network"`
	Ipv6NoTraffic    string `json:"ipv6_no_traffic"`
	Ipv6Subnet       string `json:"ipv6_subnet"`
}

type loggedInUser struct {
	Username    string `json:"username"`
	SessionName string `json:"session_name"`
	SessionId   string `json:"session_id"`
	State       string `json:"state"`
	IdleTime    string `json:"idle_time"`
	LogonTime   string `json:"logon_time"`
}

func parseConnectivity(output string) (connectivity, error) {
	var connectInfoList []connectivity
	err := json.Unmarshal([]byte(output), &connectInfoList)
	return connectInfoList[0], err
}

func (sshConnection SshConnectionInfo) GetConnectivity() (connectivity, error) {
	var connectInfo connectivity
	result, err := sshConnection.RunCommandFromSSHConnectionUseKeys(`osqueryi --json "SELECT * FROM connectivity"`)
	if err != nil {
		return connectInfo, err
	}
	connectInfo, err = parseConnectivity(result)
	return connectInfo, err
}

func parseCPUInfo(output string) (cpuInfo, error) {
	var cpuInfoList []cpuInfo
	err := json.Unmarshal([]byte(output), &cpuInfoList)
	return cpuInfoList[0], err
}

func (sshConnection SshConnectionInfo) GetCPUInfo() (cpuInfo, error) {
	var cpu cpuInfo
	result, err := sshConnection.RunCommandFromSSHConnectionUseKeys(`osqueryi --json "SELECT * FROM cpu_info"`)
	if err != nil {
		return cpu, err
	}
	cpu, err = parseCPUInfo(result)
	return cpu, err
}

func parseInterfacesInfo(output string) ([]interfaceInfo, error) {
	var interfaceList []interfaceInfo
	err := json.Unmarshal([]byte(output), &interfaceList)
	return interfaceList, err
}

func (sshConnection SshConnectionInfo) GetIntefaceList() ([]interfaceInfo, error) {
	var interfaceList []interfaceInfo
	result, err := sshConnection.RunCommandFromSSHConnectionUseKeys(`osqueryi --json "SELECT * FROM interface_details"`)
	if err != nil {
		return interfaceList, err
	}
	interfaceList, err = parseInterfacesInfo(result)
	return interfaceList, err
}

func parseOSVersion(output string) (osVersion, error) {
	var osVersionList []osVersion
	err := json.Unmarshal([]byte(output), &osVersionList)
	return osVersionList[0], err
}

func (sshConnection SshConnectionInfo) GetOSVersion() (osVersion, error) {
	var os osVersion
	result, err := sshConnection.RunCommandFromSSHConnectionUseKeys(`osqueryi --json "SELECT * FROM os_version"`)
	if err != nil {
		return os, err
	}
	os, err = parseOSVersion(result)
	return os, err
}

func ParseAnsibleFactsOutput(output string) error {
	openingTag := strings.Index(output, "{")
	if openingTag < 0 {
		return errors.New("Ansible Output is not JSON format")
	}
	closingTag := strings.LastIndex(output, "}")
	if closingTag < 0 {
		return errors.New("Ansible Output is not JSON format")
	}
	jsonStr := output[openingTag:closingTag]
	jsonStr += "}"
	jsonParsed, err := gabs.ParseJSON([]byte(jsonStr))
	if err != nil {
		return err
	}
	fmt.Println(jsonParsed.Path("ansible_facts.architecture").Data())
	fmt.Println(jsonParsed.Path("ansible_facts.windows_domain").Data())
	fmt.Println(jsonParsed.Path("ansible_facts.uptime_seconds").Data())
	fmt.Println(jsonParsed.Path("ansible_facts.hostname").Data())
	fmt.Println(jsonParsed.Path("ansible_facts.memtotal_mb").Data())
	fmt.Println(jsonParsed.Path("ansible_facts.distribution").Data())
	fmt.Println(jsonParsed.Path("ansible_facts.distribution_version").Data())
	fmt.Println(jsonParsed.Path("ansible_facts.bios_date").Data())
	fmt.Println(jsonParsed.Path("ansible_facts.bios_version").Data())
	return nil
}
