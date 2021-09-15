package models

import (
	"errors"
	"fmt"
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

type VyOsJson struct {
	NetworkId int          `json:"networkId"`
	Interface string       `json:"interface"`
	IPv4      IPv4VyosInfo `json:"ipv4"`
	IPv6      IPv6VyosInfo `json:"ipv6"`
}

func GetInfoVyos(sshConnectionId int) ([]L3_VyosInterfaces, error) {
	var interfacesList []L3_VyosInterfaces

	// Get Hostname from Id
	hostname, err := GetSshHostnameFromId(sshConnectionId)
	if err != nil {
		return interfacesList, errors.New("fail to get hostname")
	}

	interfacesList, err = GetInterfacesVyos(hostname)
	if err != nil {
		return interfacesList, errors.New("fail to get vyos interfaces information")
	}
	return interfacesList, err

}

// Get Information about IP
func GetInterfacesVyos(hostname string) ([]L3_VyosInterfaces, error) {
	var (
		interfacesList []L3_VyosInterfaces
		err            error
	)

	// Load YAML file
	var extraValue map[string]string = map[string]string{"host": hostname}
	ouput, err := LoadYAML("./yamls/vyos_getinfo.yml", extraValue)
	if err != nil {
		return interfacesList, errors.New("fail to load yamls")
	}

	// Get substring from ansible output
	data := utils.ExtractSubString(ouput, "ok: [vyos] => ", "PLAY RECAP")

	// Parse Json format
	jsonParsed, err := gabs.ParseJSON([]byte(data))
	if err != nil {
		fmt.Println(err.Error())
	}

	value, err := jsonParsed.Search("msg", "l3_interfaces").Children()
	if err != nil {
		fmt.Println(err.Error())
	}

	for _, child := range value {
		var interfaces L3_VyosInterfaces
		// Get Interface name
		if child.Exists("name") {
			interfaces.Name = child.Search("name").String()
		}
		// Get Ipv4
		if child.Exists("ipv4", "address") {
			rawString := child.Search("ipv4", "address").String()
			result := TrimStringOfIP(rawString)
			interfaces.IPv4.Address = result
		}
		// Get Ipv6
		if child.Exists("ipv6", "address") {
			rawString := child.Search("ipv6", "address").String()
			result := TrimStringOfIP(rawString)
			interfaces.IPv6.Address = result
		}
		interfacesList = append(interfacesList, interfaces)
	}

	return interfacesList, err
}

func ConfigIPVyos(vyosJson VyOsJson) (string, error) {
	var (
		output string
		err    error
	)
	// Get Hostname from Id
	hostname, err := GetSshHostnameFromId(vyosJson.NetworkId)
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
