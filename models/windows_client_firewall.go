package models

import (
	"strings"
)

type AppFirewallRule struct {
	Name                  string
	DisplayName           string
	Description           string
	Group                 string
	Enabled               bool
	Profile               string
	Platform              string
	Direction             string
	Action                string
	EdgeTraversalPolicy   string
	LooseSourceMapping    bool
	LocalOnlyMapping      bool
	Owner                 string
	PrimaryStatus         string
	Status                string
	EnforcementStatus     string
	PolicyStoreSource     string
	PolicyStoreSourceType string
}

type PortNetshFirewallRule struct {
	RuleName      string
	Enabled       string
	Direction     string
	Profiles      string
	Grouping      string
	LocalIP       string
	RemoteIP      string
	Protocol      string
	LocalPort     string
	RemotePort    string
	EdgeTraversal string
	Action        string
}

type PortFirewallRule struct {
	Protocol      string
	Localport     string
	RemotePort    string
	IcmpType      string
	DynamicTarget string
}

func ParsePortNetshFirewallRuleFromPowershell(result string) ([]PortNetshFirewallRule, error) {
	result = strings.Trim(result, "\n")
	var rule PortNetshFirewallRule
	var ruleList []PortNetshFirewallRule
	var counter int = 0
	var tmpArray [12]string
	lines := strings.Split(result, "\n")
	var line string
	for i := 0; i < len(lines); i++ {
		line = lines[i]
		if line != "" && line != "\t" && line != "\r" && !strings.Contains(line, "-----------") && line != "Ok.\r" {
			if strings.Contains(line, "ICMPv6-In") || strings.Contains(line, "ICMPv6-Out") || strings.Contains(line, "ICMPv4-In") || strings.Contains(line, "ICMPv4-Out") {
				i += 12
				counter = 0
				continue
			}
			arrays := strings.Split(line, ":")
			tmpArray[counter] = strings.TrimSpace(arrays[1])
			if counter == 7 && strings.TrimSpace(tmpArray[counter]) == "Any" {
				tmpArray[counter+1] = "Any"
				tmpArray[counter+2] = "Any"
				counter += 2
			}
			counter++
		}
		if counter == 12 {
			counter = 0
			rule.RuleName = tmpArray[0]
			rule.Enabled = tmpArray[1]
			rule.Direction = tmpArray[2]
			rule.Profiles = tmpArray[3]
			rule.Grouping = tmpArray[4]
			rule.LocalIP = tmpArray[5]
			rule.RemoteIP = tmpArray[6]
			rule.Protocol = tmpArray[7]
			rule.LocalPort = tmpArray[8]
			rule.RemotePort = tmpArray[9]
			rule.EdgeTraversal = tmpArray[10]
			rule.Action = tmpArray[11]
			ruleList = append(ruleList, rule)
		}
	}
	return ruleList, nil
}

type AppliedFirewallRule struct {
	Host       []string `json:"host"`
	RuleName   string   `json:"rule_name"`
	Enabled    string   `json:"enabled"`
	Direction  string   `json:"direction"`
	Profiles   []string `json:"profiles"`
	Grouping   string   `json:"group"`
	LocalIP    string   `json:"local_ip"`
	RemoteIP   string   `json:"remote_ip"`
	Protocol   string   `json:"protocol"`
	LocalPort  string   `json:"local_port"`
	RemotePort string   `json:"remote_port"`
	Action     string   `json:"rule_action"`
}

type DeletedFirewallRule struct {
	Host       string `json:"host"`
	RuleName   string `json:"rule_name"`
	Enabled    string `json:"enabled"`
	Direction  string `json:"direction"`
	Profiles   string `json:"profiles"`
	Grouping   string `json:"group"`
	LocalIP    string `json:"local_ip"`
	RemoteIP   string `json:"remote_ip"`
	Protocol   string `json:"protocol"`
	LocalPort  string `json:"local_port"`
	RemotePort string `json:"remote_port"`
	Action     string `json:"rule_action"`
}

func AddFirewallRule(firewallJson string) (string, error) {
	output, err := RunAnsiblePlaybookWithjson("./yamls/windows_client/add_firewall_rule.yml", firewallJson)
	return output, err
}

func DeleteFirewallRule(ruleNameJson string) (string, error) {
	output, err := RunAnsiblePlaybookWithjson("./yamls/windows_client/delete_firewall_rule.yml", ruleNameJson)
	return output, err
}

func ParsePortFirewallRuleFromPowershell(result string) ([]PortFirewallRule, error) {
	result = strings.Trim(result, "\n")
	var rule PortFirewallRule
	var ruleList []PortFirewallRule
	var counter int = 0
	var tmpArray [5]string
	for _, line := range strings.Split(result, "\n") {
		if line != "" && line != "\t" && line != "\r" {
			arrays := strings.Split(line, ":")
			tmpArray[counter] = strings.TrimSpace(arrays[1])
			counter++
		}
		if counter == 5 {
			counter = 0
			rule.Protocol = tmpArray[0]
			rule.Localport = tmpArray[1]
			rule.RemotePort = tmpArray[2]
			rule.IcmpType = tmpArray[3]
			rule.DynamicTarget = tmpArray[4]
			ruleList = append(ruleList, rule)
		}
	}
	return ruleList, nil
}
