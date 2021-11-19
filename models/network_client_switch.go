package models

import (
	"encoding/json"
	"errors"
	"regexp"
	"strconv"
	"strings"
)

type SwitchJson struct {
	SshConnectionId []int    `json:"sshConnectionId"`
	Host            string   `json:"host"`
	VlanId          int      `json:"vlanId"`
	VlanName        string   `json:"vlanName"`
	Interfaces      []string `json:"interfaces"`
}

type SwitchVlan struct {
	VlanId     int    `json:"vlanId"`
	VlanName   string `json:"vlanName"`
	VlanStatus string `json:"vlanStatus"`
	VlanPorts  string `json:"vlanPorts"`
}

// create vlan
func CreateVlanSwitch(switchJson SwitchJson) ([]string, error) {
	var (
		outputList []string
		err        error
	)
	// Get Hostname from Id
	for _, id := range switchJson.SshConnectionId {
		sshConnection, err := GetSSHConnectionFromId(id)
		if err != nil {
			return outputList, errors.New("fail to parse id")
		}

		switchJson.Host = sshConnection.HostNameSSH

		// Marshal and run playbook
		switchJsonMarshal, err := json.Marshal(switchJson)
		if err != nil {
			return outputList, err
		}
		var filepath string
		if sshConnection.NetworkOS == "ios" {
			filepath = "./yamls/network_client/cisco/cisco_switch_config_createvlan.yml"
		}
		output, err := RunAnsiblePlaybookWithjson(filepath, string(switchJsonMarshal))
		if err != nil {
			return outputList, errors.New("fail to load yaml file")
		}
		outputList = append(outputList, output)
	}
	return outputList, err
}

// get vlan switch
func GetVlanSwitch(sshConnectionId int) ([]SwitchVlan, error) {
	var (
		switchVlanList []SwitchVlan
		err            error
	)
	sshConnection, err := GetSSHConnectionFromId(sshConnectionId)
	if err != nil {
		return switchVlanList, errors.New("fail to get ssh connection")
	}

	if sshConnection.NetworkOS == "junos" || sshConnection.NetworkOS == "vyos" {
		return switchVlanList, errors.New("function is not support on the device")
	}

	if sshConnection.NetworkOS == "ios" {
		switchVlanList, err = ParseVlanSwitchCisco(*sshConnection)
	}
	if err != nil {
		return switchVlanList, errors.New("fail to get vlan information")
	}

	return switchVlanList, err
}

func ParseVlanSwitchCisco(sshConnection SshConnectionInfo) ([]SwitchVlan, error) {
	var (
		switchVlanList []SwitchVlan
		err            error
	)

	output, err := sshConnection.RunCommandFromSSHConnectionUseKeys("show vlan brief")
	if err != nil {
		return switchVlanList, errors.New("fail to run remote command")
	}

	// spilit output
	lines := strings.Split(output, "\n")
	r, _ := regexp.Compile("^[0-9]+")
	for index, line := range lines {
		// Skip header line
		if index > 3 {
			var switchVlan SwitchVlan
			match := r.MatchString(line)
			if match {
				// Get VlanId
				rawVlanId := strings.TrimSpace(line[:4])
				switchVlan.VlanId, err = strconv.Atoi(rawVlanId)
				if err != nil {
					return switchVlanList, errors.New("fail to convert string to id")
				}

				// Get VlanName
				switchVlan.VlanName = strings.TrimSpace(line[5:37])

				// Get VlanStatus
				switchVlan.VlanStatus = strings.TrimSpace(line[38:47])

				// Get VlanPorts
				switchVlan.VlanPorts = strings.TrimSpace(line[48:])

				// Append to list
				switchVlanList = append(switchVlanList, switchVlan)
			} else {
				vlanPorts := strings.TrimSpace(line)
				indexSwitchVlanLatest := len(switchVlanList) - 1

				// Update ports
				switchVlanList[indexSwitchVlanLatest].VlanPorts += "," + vlanPorts
			}
		}

	}
	return switchVlanList, err
}

// add interfaces to vlan
func AddInterfacesToVlanSwitch(switchJson SwitchJson) ([]string, error) {
	var (
		outputList []string
		err        error
	)
	// Get Hostname from Id
	for _, id := range switchJson.SshConnectionId {
		sshConnection, err := GetSSHConnectionFromId(id)
		if err != nil {
			return outputList, errors.New("fail to parse id")
		}
		if sshConnection.NetworkOS == "junos" || sshConnection.NetworkOS == "vyos" {
			return outputList, errors.New("function is not supported on the device")
		}

		switchJson.Host = sshConnection.HostNameSSH

		// Marshal and run playbook
		switchJsonMarshal, err := json.Marshal(switchJson)
		if err != nil {
			return outputList, err
		}
		var filepath string
		if sshConnection.NetworkOS == "ios" {
			filepath = "./yamls/network_client/cisco/cisco_switch_config_interfacetovlan.yml"
		}
		output, err := RunAnsiblePlaybookWithjson(filepath, string(switchJsonMarshal))
		if err != nil {
			return outputList, errors.New("fail to load yaml file")
		}
		outputList = append(outputList, output)
	}
	return outputList, err
}

// delete vlan
func DeleteVlanSwitch(switchJson SwitchJson) ([]string, error) {
	var (
		outputList []string
		err        error
	)
	// Get Hostname from Id
	for _, id := range switchJson.SshConnectionId {
		sshConnection, err := GetSSHConnectionFromId(id)
		if err != nil {
			return outputList, errors.New("fail to parse id")
		}

		switchJson.Host = sshConnection.HostNameSSH

		// Marshal and run playbook
		switchJsonMarshal, err := json.Marshal(switchJson)
		if err != nil {
			return outputList, err
		}
		var filepath string
		if sshConnection.NetworkOS == "ios" {
			filepath = "./yamls/network_client/cisco/cisco_switch_config_deletevlan.yml"
		}
		output, err := RunAnsiblePlaybookWithjson(filepath, string(switchJsonMarshal))
		if err != nil {
			return outputList, errors.New("fail to load yaml file")
		}
		outputList = append(outputList, output)
	}
	return outputList, err
}

// get vlan switch
func GetInterfaceSwitch(sshConnectionId int) ([]string, error) {
	var (
		InterfaceNameList []string
		err               error
	)

	interfaceList, err := GetInfoInterfaceCisco(sshConnectionId)
	for _, interfaces := range interfaceList {
		if !strings.Contains(interfaces.Name, "Null") && !strings.Contains(interfaces.Name, "Vlan") {
			InterfaceNameList = append(InterfaceNameList, interfaces.Name)
		}
	}

	return InterfaceNameList, err
}
