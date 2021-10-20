package models

import (
	"encoding/json"
	"errors"
	"net"
	"strconv"
)

type RouterJson struct {
	SshConnectionId []int  `json:"sshConnectionId"`
	Host            string `json:"host"`
	Interface       string `json:"interface"`
	Address4        string `json:"address4"`
	NetMask4        string `json:"netmask4"`
	Address6        string `json:"address6"`
	Prefix          string `json:"prefix"`
	Mask            string `json:"mask"`
	Next_hop        string `json:"next_hop"`
	Dest            string `json:"dest"`
	Enabled         bool   `json:"enabled"`
}

// config ipv4 and ipv6
func ConfigIPRouter(routerJson RouterJson) ([]string, error) {
	var (
		outputList []string
		err        error
	)
	// Get Hostname from Id
	for _, id := range routerJson.SshConnectionId {
		sshConnection, err := GetSSHConnectionFromId(id)
		if err != nil {
			return outputList, errors.New("fail to parse id")
		}

		routerJson.Host = sshConnection.HostNameSSH

		// Convert netmask to prefix length
		stringMask := net.IPMask(net.ParseIP(routerJson.NetMask4).To4())

		length, _ := stringMask.Size()
		routerJson.Address4 += "/" + strconv.Itoa(length)

		// Marshal and run playbook
		ciscoJsonMarshal, err := json.Marshal(routerJson)
		if err != nil {
			return outputList, err
		}
		var filepath string
		if sshConnection.NetworkOS == "ios" {
			filepath = "./yamls/network_client/cisco/cisco_config_ip.yml"
		} else if sshConnection.NetworkOS == "vyos" {
			filepath = "./yamls/network_client/vyos/vyos_config_ip.yml"
		} else if sshConnection.NetworkOS == "junos" {
			filepath = "./yamls/network_client/juniper/juniper_config_ip.yml"
		}
		output, err := RunAnsiblePlaybookWithjson(filepath, string(ciscoJsonMarshal))
		if err != nil {
			return outputList, errors.New("fail to load yaml file")
		}
		outputList = append(outputList, output)
	}
	return outputList, err
}
