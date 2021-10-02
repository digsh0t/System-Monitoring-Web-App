package models

import (
	"encoding/json"
	"errors"
	"sort"
	"strconv"
	"strings"

	"github.com/Jeffail/gabs"
	"github.com/wintltr/login-api/utils"
)

type IPv4CiscoInfo struct {
	Address string `json:"address"`
	Subnet  string `json:"subnet"`
}
type IPv6CiscoInfo struct {
	Address string `json:"address"`
	Subnet  string `json:"subnet"`
}

type L3_CiscoInterfaces struct {
	Name         string        `json:"name"`
	BandWidth    int           `json:"bandwidth"`
	Description  string        `json:"description"`
	Duplex       string        `json:"duplex"`
	IPv4         IPv4CiscoInfo `json:"ipv4"`
	IPv6         IPv6CiscoInfo `json:"ipv6"`
	Lineprotocol string        `json:"lineprotocol"`
	Macaddress   string        `json:"macaddress"`
	Mtu          int           `json:"mtu"`
	Operstatus   string        `json:"operstatus"`
	Type         string        `json:"type"`
}

type CiscoJson struct {
	SshConnectionId []int    `json:"sshConnectionId"`
	Host            []string `json:"host"`
	Interface       string   `json:"interface"`
	Address4        string   `json:"address4"`
	Address6        string   `json:"address6"`
	Prefix          string   `json:"prefix"`
	Mask            string   `json:"mask"`
	Next_hop        string   `json:"next_hop"`
	Dest            string   `json:"dest"`
}

func GetInfoConfigCisco(sshConnectionId int) ([]string, error) {
	var configCisco []string

	// Get Hostname from Id
	hostname, err := GetSshHostnameFromId(sshConnectionId)
	if err != nil {
		return configCisco, errors.New("fail to get hostname")
	}

	ciscoJson := CiscoJson{
		Host: []string{hostname},
	}

	// Marshal
	ciscoJsonMarshal, err := json.Marshal(ciscoJson)
	if err != nil {
		return configCisco, errors.New("fail to marshal json")
	}

	// Load YAML
	output, err := RunAnsiblePlaybookWithjson("./yamls/network_client/cisco/cisco_getconfig.yml", string(ciscoJsonMarshal))
	if err != nil {
		return configCisco, errors.New("fail to run playbook")
	}

	// Get substring from ansible output
	data := utils.ExtractSubString(output, " => ", "PLAY RECAP")

	// Parse Json format
	jsonParsed, err := gabs.ParseJSON([]byte(data))
	if err != nil {
		return configCisco, err
	}

	// Get msg and spilit
	rawData := strings.TrimSpace(jsonParsed.Search("msg").String())
	rawDataList := strings.Split(rawData, "\\n")

	// Return
	configCisco = append(configCisco, rawDataList...)

	return configCisco, err
}

func GetInfoInterfaceCisco(sshConnectionId int) ([]L3_CiscoInterfaces, error) {
	var l3_interfacesList []L3_CiscoInterfaces

	// Get Hostname from Id
	hostname, err := GetSshHostnameFromId(sshConnectionId)
	if err != nil {
		return l3_interfacesList, errors.New("fail to get hostname")
	}

	ciscoJson := CiscoJson{
		Host: []string{hostname},
	}

	// Marshal
	ciscoJsonMarshal, err := json.Marshal(ciscoJson)
	if err != nil {
		return l3_interfacesList, errors.New("fail to marshal json")
	}

	// Load YAML
	output, err := RunAnsiblePlaybookWithjson("./yamls/network_client/cisco/cisco_getinterface.yml", string(ciscoJsonMarshal))
	if err != nil {
		return l3_interfacesList, errors.New("fail to run playbook")
	}

	// Get substring from ansible output
	data := utils.ExtractSubString(output, " => ", "PLAY RECAP")

	// Parse Json format
	jsonParsed, err := gabs.ParseJSON([]byte(data))
	if err != nil {
		return l3_interfacesList, err
	}

	list, err := jsonParsed.Search("msg").ChildrenMap()
	if err != nil {
		return l3_interfacesList, errors.New("fail to parse json")
	}

	// Sort again name of interfaces
	var stringSortList []string
	for nameInterface, _ := range list {
		stringSortList = append(stringSortList, nameInterface)
	}
	sort.Strings(stringSortList)

	// Get interface name from ansible output
	for _, nameInterface := range stringSortList {
		var l3_interfaces L3_CiscoInterfaces

		// Get name interface
		l3_interfaces.Name = nameInterface

		// Get bandwith
		if jsonParsed.Exists("msg", nameInterface, "bandwidth") {
			result := jsonParsed.Search("msg", nameInterface, "bandwidth").String()
			l3_interfaces.BandWidth, err = strconv.Atoi(result)
			if err != nil {
				return l3_interfacesList, err
			}
		}

		// Get Description
		if jsonParsed.Exists("msg", nameInterface, "description") {
			result := jsonParsed.Search("msg", nameInterface, "description").String()
			l3_interfaces.Description = result
		}

		// Get Duplex
		if jsonParsed.Exists("msg", nameInterface, "duplex") {
			result := jsonParsed.Search("msg", nameInterface, "duplex").String()
			l3_interfaces.Duplex = result
		}

		// Get Ipv4 address and subnet
		if jsonParsed.Exists("msg", nameInterface, "ipv4", "address") {
			result := jsonParsed.Search("msg", nameInterface, "ipv4", "address").String()
			l3_interfaces.IPv4.Address = result
		}

		if jsonParsed.Exists("msg", nameInterface, "ipv4", "subnet") {
			result := jsonParsed.Search("msg", nameInterface, "ipv4", "subnet").String()
			l3_interfaces.IPv4.Subnet = result
		}

		// Get Ipv6 address and subnet
		if jsonParsed.Exists("msg", nameInterface, "ipv6", "address") {
			result := jsonParsed.Search("msg", nameInterface, "ipv6", "address").String()
			l3_interfaces.IPv6.Address = result
		}

		if jsonParsed.Exists("msg", nameInterface, "ipv6", "subnet") {
			result := jsonParsed.Search("msg", nameInterface, "ipv6", "subnet").String()
			l3_interfaces.IPv6.Subnet = result
		}

		// Get Line Protocol
		if jsonParsed.Exists("msg", nameInterface, "lineprotocol") {
			result := jsonParsed.Search("msg", nameInterface, "lineprotocol").String()
			l3_interfaces.Lineprotocol = result
		}

		// Get Mac Address
		if jsonParsed.Exists("msg", nameInterface, "macaddress") {
			result := jsonParsed.Search("msg", nameInterface, "macaddress").String()
			l3_interfaces.Macaddress = result
		}

		// Get MTU
		if jsonParsed.Exists("msg", nameInterface, "mtu") {
			result := jsonParsed.Search("msg", nameInterface, "mtu").String()
			l3_interfaces.Mtu, err = strconv.Atoi(result)
			if err != nil {
				return l3_interfacesList, err
			}
		}

		// Get Operate Status
		if jsonParsed.Exists("msg", nameInterface, "operstatus") {
			result := jsonParsed.Search("msg", nameInterface, "operstatus").String()
			l3_interfaces.Operstatus = result
		}

		// Get Type
		if jsonParsed.Exists("msg", nameInterface, "type") {
			result := jsonParsed.Search("msg", nameInterface, "type").String()
			l3_interfaces.Type = result
		}

		l3_interfacesList = append(l3_interfacesList, l3_interfaces)
	}

	return l3_interfacesList, err
}

// config ipv4 and ipv6
func ConfigIPCisco(ciscoJson CiscoJson) (string, error) {
	var (
		output string
		err    error
	)
	// Get Hostname from Id
	var host []string
	for _, id := range ciscoJson.SshConnectionId {
		hostname, err := GetSSHConnectionFromId(id)
		if err != nil {
			return output, errors.New("fail to parse id")
		}
		host = append(host, hostname.HostNameSSH)
	}
	ciscoJson.Host = host

	ciscoJsonMarshal, err := json.Marshal(ciscoJson)
	if err != nil {
		return output, err
	}

	output, err = RunAnsiblePlaybookWithjson("./yamls/network_client/cisco/cisco_config_ip.yml", string(ciscoJsonMarshal))
	if err != nil {
		return output, errors.New("fail to load yaml file")
	}
	return output, err
}

// confic Static route
func ConfigStaticRouteCisco(ciscoJson CiscoJson) (string, error) {
	var (
		output string
		err    error
	)
	// Get Hostname from Id
	var host []string
	for _, id := range ciscoJson.SshConnectionId {
		hostname, err := GetSSHConnectionFromId(id)
		if err != nil {
			return output, errors.New("fail to parse id")
		}
		host = append(host, hostname.HostNameSSH)
	}
	ciscoJson.Host = host

	ciscoJsonMarshal, err := json.Marshal(ciscoJson)
	if err != nil {
		return output, err
	}

	output, err = RunAnsiblePlaybookWithjson("./yamls/network_client/cisco/cisco_config_staticroute.yml", string(ciscoJsonMarshal))
	if err != nil {
		return output, errors.New("fail to load yaml file")
	}
	return output, err
}

// Test Ping
func TestPingCisco(ciscoJson CiscoJson) (string, error) {
	var (
		output string
		err    error
	)
	// Get Hostname from Id
	var host []string
	for _, id := range ciscoJson.SshConnectionId {
		hostname, err := GetSSHConnectionFromId(id)
		if err != nil {
			return output, errors.New("fail to parse id")
		}
		host = append(host, hostname.HostNameSSH)
	}
	ciscoJson.Host = host

	ciscoJsonMarshal, err := json.Marshal(ciscoJson)
	if err != nil {
		return output, err
	}

	output, err = RunAnsiblePlaybookWithjson("./yamls/network_client/cisco/cisco_test_ping.yml", string(ciscoJsonMarshal))
	if err != nil {
		return output, errors.New("fail to load yaml file")
	}
	return output, err
}
