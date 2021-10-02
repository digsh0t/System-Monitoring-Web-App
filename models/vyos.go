package models

import (
	"encoding/json"
	"errors"
	"fmt"
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
	Host       []int    `json:"host"`
	HosrString []string `json:"host_var"`
	Interface  string   `json:"interface"`
	Address4   string   `json:"address4"`
	Address6   string   `json:"address6"`
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

// Get Information About Vyos's IP
func GetInterfacesVyos(hostname string) (VyosInfo, error) {

	// Declare variables
	var (
		vyOSInfo VyosInfo
		err      error
	)

	// Append value for json object
	vyosJson := VyOsJson{
		HosrString: []string{hostname},
	}

	// Marshal json for running playbook
	vyosJsonMarshal, err := json.Marshal(vyosJson)
	if err != nil {
		return vyOSInfo, err
	}

	// Load YAML file
	ouput, err := RunAnsiblePlaybookWithjson("./yamls/network_client/vyos/vyos_getinfo.yml", string(vyosJsonMarshal))
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
	var host []string
	for _, id := range vyosJson.Host {
		hostname, err := GetSSHConnectionFromId(id)
		if err != nil {
			return output, errors.New("fail to parse id")
		}
		host = append(host, hostname.HostNameSSH)
	}
	vyosJson.HosrString = host

	vyosJsonMarshal, err := json.Marshal(vyosJson)
	if err != nil {
		return output, err
	}

	output, err = RunAnsiblePlaybookWithjson("./yamls/network_client/vyos/vyos_config_ip.yml", string(vyosJsonMarshal))
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

// Correct format for IP
func TrimConfig(s string) string {
	s = strings.TrimRight(s, `,"m[b10u\`)
	return s
}

func GetInfoConfigVyos(sshConnectionId int) ([]string, error) {
	var configVyos []string

	// Get Hostname from Id
	hostname, err := GetSshHostnameFromId(sshConnectionId)
	if err != nil {
		return configVyos, errors.New("fail to get hostname")
	}

	vyosJson := VyOsJson{
		HosrString: []string{hostname},
	}

	// Marshal
	vyosJsonMarshal, err := json.Marshal(vyosJson)
	if err != nil {
		return configVyos, errors.New("fail to marshal json")
	}

	// Load YAML
	output, err := RunAnsiblePlaybookWithjson("./yamls/network_client/vyos/vyos_getconfig.yml", string(vyosJsonMarshal))
	if err != nil {
		return configVyos, errors.New("fail to run playbook")
	}

	// Get substring from ansible output
	data := utils.ExtractSubString(output, " => ", "PLAY RECAP")

	// Parse Json format
	jsonParsed, err := gabs.ParseJSON([]byte(data))
	if err != nil {
		return configVyos, err
	}

	// Get msg and spilit
	rawData := jsonParsed.Search("msg").String()
	fmt.Println(rawData)
	if err != nil {
		return configVyos, err
	}
	list := strings.Split(rawData, "\n")
	fmt.Println("len", len(list))
	return configVyos, err
}
