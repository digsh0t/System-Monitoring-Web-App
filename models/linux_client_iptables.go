package models

import "encoding/json"

type IptableRule struct {
	Bytes        string `json:"bytes"`
	Chain        string `json:"chain"`
	DstIP        string `json:"dst_ip"`
	DstMask      string `json:"dst_mask"`
	DstPort      string `json:"dst_port"`
	FilterName   string `json:"filter_name"`
	Iniface      string `json:"iniface"`
	InifaceMask  string `json:"iniface_mask"`
	Match        string `json:"match"`
	Outiface     string `json:"outiface"`
	OutifaceMask string `json:"outiface_mask"`
	Packets      string `json:"packets"`
	Policy       string `json:"policy"`
	Protocol     string `json:"protocol"`
	SrcIP        string `json:"src_ip"`
	SrcMask      string `json:"src_mask"`
	SrcPort      string `json:"src_port"`
	Target       string `json:"target"`
}

func ParseIptables(cmdResult string) ([]IptableRule, error) {
	var Iptables []IptableRule
	err := json.Unmarshal([]byte(cmdResult), &Iptables)
	return Iptables, err
}
