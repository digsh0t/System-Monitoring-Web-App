package models

import (
	"log"
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

type PortFirewallRule struct {
	Protocol      string
	Localport     string
	RemotePort    string
	IcmpType      string
	DynamicTarget string
}

func ParsePortFirewallRuleFromPowershell(result string) ([]PortFirewallRule, error) {
	result = strings.Trim(result, "\n")
	var rule PortFirewallRule
	var ruleList []PortFirewallRule
	var counter int = 0
	var tmpArray [5]string
	for _, line := range strings.Split(result, "\n") {
		if line != "" && line != "\t" {
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
	log.Println(ruleList)
	return ruleList, nil
}
