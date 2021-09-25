package models

import (
	"encoding/json"
	"errors"
)

type IptablesRule struct {
	FilterName string `json:"filter_name"`
	Policy     string `json:"policy"`
	Protocol   string `json:"protocol"`
	SrcIP      string `json:"src_ip"`
	DstIP      string `json:"dst_ip"`
	Chain      string `json:"chain"`
	Target     string `json:"target"`
}

type IptablesJson struct {
	SshConnectionId []int    `json:"sshConnectionId"`
	Host            []string `json:"host"`
	Chain           string   `json:"chain"`
	SourceIP        string   `json:"src_ip"`
	Destination     string   `json:"dst_ip"`
	Protocol        string   `json:"protocol"`
	Jump            string   `json:"target"`
}

func ParseIptables(cmdResult string) ([]IptablesRule, error) {
	var Iptables []IptablesRule
	err := json.Unmarshal([]byte(cmdResult), &Iptables)
	return Iptables, err
}

func LinuxClientIptablesListAll(sshConnectionId int) ([]IptablesRule, error) {
	var (
		clientIptablesList []IptablesRule
		result             string
	)
	SshConnectionInfo, err := GetSSHConnectionFromId(sshConnectionId)
	if err != nil {
		return clientIptablesList, errors.New("fail to get client connection")
	}

	result, err = SshConnectionInfo.RunCommandFromSSHConnectionUseKeys(`osqueryi --json "SELECT * FROM iptables"`)
	if err != nil {
		return clientIptablesList, errors.New("fail to get clientiptables")
	}
	err = json.Unmarshal([]byte(result), &clientIptablesList)
	if err != nil {
		return clientIptablesList, errors.New("fail to get client iptables")
	}

	return clientIptablesList, nil
}

// Add iptables rule for clients
func LinuxClientIptablesAdd(iptablesJson IptablesJson) (string, error) {
	var (
		output string
		err    error
	)

	var host []string
	for _, id := range iptablesJson.SshConnectionId {
		sshConnection, err := GetSSHConnectionFromId(id)
		if err != nil {
			return output, errors.New("fail to get list connection")
		}
		host = append(host, sshConnection.HostNameSSH)
	}
	iptablesJson.Host = host

	iptablesJsonMarshal, err := json.Marshal(iptablesJson)
	if err != nil {
		return output, errors.New("fail to marshal json")
	}
	output, err = RunAnsiblePlaybookWithjson("./yamls/linux_client/add_client_iptables.yml", string(iptablesJsonMarshal))
	if err != nil {
		return output, err
	}
	return output, err

}

func LinuxClientIptablesRemove(iptablesJson IptablesJson) (string, error) {
	var (
		output string
		err    error
	)

	var host []string
	for _, id := range iptablesJson.SshConnectionId {
		sshConnection, err := GetSSHConnectionFromId(id)
		if err != nil {
			return output, errors.New("fail to get user connection")
		}
		host = append(host, sshConnection.HostNameSSH)
	}
	iptablesJson.Host = host
	iptablesJsonMarshal, err := json.Marshal(iptablesJson)
	if err != nil {
		return output, errors.New("fail to marshal json")
	}
	output, err = RunAnsiblePlaybookWithjson("./yamls/linux_client/remove_client_iptables.yml", string(iptablesJsonMarshal))
	if err != nil {
		return output, err
	}
	return output, nil

}
