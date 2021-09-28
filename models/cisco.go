package models

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/Jeffail/gabs"
	"github.com/wintltr/login-api/utils"
)

type IPv4CiscoInfo struct {
	Address string `json:"address"`
	Subnet  string `json:"subnet"`
}

type L3_CiscoInterfaces struct {
	Name         string        `json:"name"`
	BandWidth    int           `json:"bandwidth"`
	Description  string        `json:"description"`
	Duplex       string        `json:"duplex"`
	IPv4         IPv4CiscoInfo `json:"ipv4"`
	Lineprotocol string        `json:"lineprotocol"`
	Macaddress   string        `json:"macadress"`
	Mtu          int           `json:"mtu"`
	Operstatus   string        `json:"operstatus"`
	Type         string        `json:"type"`
}

type CiscoJson struct {
	SshConnectionId []int    `json:"sshConnectionId"`
	Host            []string `json:"host"`
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

	// Get interface name from ansible output
	for key, _ := range list {

		var l3_interfaces L3_CiscoInterfaces

		// Get name interface
		l3_interfaces.Name = key

		// Get bandwith
		if jsonParsed.Exists("msg", key, "bandwidth") {
			result := jsonParsed.Search("msg", key, "bandwidth").String()
			l3_interfaces.BandWidth, err = strconv.Atoi(result)
			if err != nil {
				return l3_interfacesList, err
			}
		}

		// Get Description
		if jsonParsed.Exists("msg", key, "description") {
			result := jsonParsed.Search("msg", key, "description").String()
			l3_interfaces.Description = result
		}

		// Get Duplex
		if jsonParsed.Exists("msg", key, "duplex") {
			result := jsonParsed.Search("msg", key, "duplex").String()
			l3_interfaces.Duplex = result
		}

		// Get Ipv4 address and subnet
		if jsonParsed.Exists("msg", key, "ipv4", "address") {
			result := jsonParsed.Search("msg", key, "ipv4", "address").String()
			l3_interfaces.IPv4.Address = result
		}

		if jsonParsed.Exists("msg", key, "ipv4", "subnet") {
			result := jsonParsed.Search("msg", key, "ipv4", "subnet").String()
			l3_interfaces.IPv4.Subnet = result
		}

		// Get Line Protocol
		if jsonParsed.Exists("msg", key, "lineprotocol") {
			result := jsonParsed.Search("msg", key, "lineprotocol").String()
			l3_interfaces.Lineprotocol = result
		}

		// Get Mac Address
		if jsonParsed.Exists("msg", key, "macaddress") {
			result := jsonParsed.Search("msg", key, "macaddress").String()
			l3_interfaces.Macaddress = result
		}

		// Get MTU
		if jsonParsed.Exists("msg", key, "mtu") {
			result := jsonParsed.Search("msg", key, "mtu").String()
			l3_interfaces.Mtu, err = strconv.Atoi(result)
			if err != nil {
				return l3_interfacesList, err
			}
		}

		// Get Operate Status
		if jsonParsed.Exists("msg", key, "operstatus") {
			result := jsonParsed.Search("msg", key, "operstatus").String()
			l3_interfaces.Operstatus = result
		}

		// Get Type
		if jsonParsed.Exists("msg", key, "type") {
			result := jsonParsed.Search("msg", key, "type").String()
			l3_interfaces.Type = result
		}

		l3_interfacesList = append(l3_interfacesList, l3_interfaces)
	}

	return l3_interfacesList, err
}
