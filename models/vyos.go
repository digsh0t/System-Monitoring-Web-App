package models

import (
	"errors"
	"strconv"
	"strings"

	"github.com/Jeffail/gabs"
	"github.com/wintltr/login-api/utils"
)

type IPv4VyosInfo struct {
	Address string `json:"address"`
}

type IPv6VyosInfo struct {
	Address string `json:"address"`
}

type L3_VyosInterfaces struct {
	Name string       `json:"name"`
	IPv4 IPv4VyosInfo `json:"ipv4"`
	IPv6 IPv6VyosInfo `json:"ipv6"`
}
type VyosInterfaces struct {
	Name        string `json:"name"`
	Enabled     bool   `json:"enabled"`
	Description string `json:"description"`
}

type VyosInfo struct {
	Interfaces    []VyosInterfaces    `json:"interfaces"`
	L3_Interfaces []L3_VyosInterfaces `json:"l3_interfaces"`
}

type VyOsJson struct {
	Host      []int        `json:"host"`
	Interface string       `json:"interface"`
	IPv4      IPv4VyosInfo `json:"ipv4"`
	IPv6      IPv6VyosInfo `json:"ipv6"`
}

func GetInfoVyos(sshConnectionId int) (VyosInfo, error) {
	var vyosInfo VyosInfo

	// Get Hostname from Id
	hostname, err := GetSshHostnameFromId(sshConnectionId)
	if err != nil {
		return vyosInfo, errors.New("fail to get hostname")
	}

	vyosInfo, err = GetInterfacesVyos(hostname)
	if err != nil {
		return vyosInfo, errors.New("fail to get vyos interfaces information")
	}
	return vyosInfo, err

}

// Get Information about IP
func GetInterfacesVyos(hostname string) (VyosInfo, error) {
	var (
		vyOSInfo VyosInfo
		err      error
	)

	// Load YAML file
	var extraValue map[string]string = map[string]string{"host": hostname}
	ouput, err := LoadYAML("./yamls/vyos_getinfo.yml", extraValue)
	if err != nil {
		return vyOSInfo, errors.New("fail to load yamls")
	}

	// Get substring from ansible output
	data := utils.ExtractSubString(ouput, " => ", "PLAY RECAP")

	// Parse Json format
	jsonParsed, err := gabs.ParseJSON([]byte(data))
	if err != nil {
		return vyOSInfo, err
	}

	// Get Interfaces
	value, err := jsonParsed.Search("msg", "interfaces").Children()
	var interfacesList []VyosInterfaces
	if err != nil {
		return vyOSInfo, err
	}
	for _, child := range value {
		var interfaces VyosInterfaces
		if child.Exists("name") {
			result := strings.Trim(child.Search("name").String(), "\"")
			interfaces.Name = result
		}
		if child.Exists("enabled") {
			result := child.Search("enabled").String()
			interfaces.Enabled, err = strconv.ParseBool(result)
			if err != nil {
				return vyOSInfo, err
			}
		}
		if child.Exists("description") {
			result := strings.Trim(child.Search("description").String(), "\"")
			interfaces.Description = result
		}
		interfacesList = append(interfacesList, interfaces)
	}

	// Get L3_Intefaces
	value, err = jsonParsed.Search("msg", "l3_interfaces").Children()
	var l3_interfacesList []L3_VyosInterfaces
	if err != nil {
		return vyOSInfo, err
	}

	for _, child := range value {
		var l3_interfaces L3_VyosInterfaces
		// Get Interface name
		if child.Exists("name") {
			result := strings.Trim(child.Search("name").String(), "\"")
			l3_interfaces.Name = result
		}
		// Get Ipv4
		if child.Exists("ipv4", "address") {
			rawString := child.Search("ipv4", "address").String()
			result := TrimStringOfIP(rawString)
			l3_interfaces.IPv4.Address = result
		}
		// Get Ipv6
		if child.Exists("ipv6", "address") {
			rawString := child.Search("ipv6", "address").String()
			result := TrimStringOfIP(rawString)
			l3_interfaces.IPv6.Address = result
		}
		l3_interfacesList = append(l3_interfacesList, l3_interfaces)
	}

	// Append to vyOSInfo
	vyOSInfo.Interfaces = interfacesList
	vyOSInfo.L3_Interfaces = l3_interfacesList
	return vyOSInfo, err
}

func ConfigIPVyos(vyosJson VyOsJson) (string, error) {
	var (
		output string
		err    error
	)
	// Get Hostname from Id
	hostname, err := ConvertListIdToHostnameVer2(vyosJson.Host)
	if err != nil {
		return output, errors.New("fail to get hostname")
	}
	var extraValue map[string]string = map[string]string{"host": hostname, "interface": vyosJson.Interface, "address4": vyosJson.IPv4.Address, "address6": vyosJson.IPv6.Address}
	output, err = LoadYAML("./yamls/vyos_config_ip.yml", extraValue)
	if err != nil {
		return output, errors.New("fail to load yaml file")
	}
	return output, err
}

// Correct format for IP
func TrimStringOfIP(s string) string {
	s = strings.TrimLeft(s, "[\"")
	s = strings.TrimRight(s, "]\"m[b10u\\'")
	return s
}
