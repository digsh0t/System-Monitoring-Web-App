package models

import (
	"os/exec"
	"regexp"
	"strings"

	"github.com/wintltr/login-api/utils"
)

type IPv4Info struct {
	Address string `json:"address4"`
	Netmask string `json:"netmask"`
	Gateway string `json:"gateway4"`
}

type IPv6Info struct {
	Address string `json:"address6"`
	Gateway string `json:"gateway6"`
}

type Interfaces struct {
	InterfacesName string   `json:interfacename`
	Ipv4           IPv4Info `json:"ipv4"`
	Ipv6           IPv6Info `json:"ipv6"`
}

type Host struct {
	ClientName  string       `json:"clientName"`
	Interfaces  []Interfaces `json:"default_interface"`
	IsConnected bool         `json:"isconnected"`
}

// Input a text and array of keys that need to find
// Response to the array of values corresponding to keys

func extractJsonValue(ansibleOutput string, keys []string) []string {
	var valueList []string
	var isFound bool = false

	for _, keyword := range keys {

		for _, line := range strings.Split(strings.TrimRight(ansibleOutput, "\n"), "\n") {
			line = strings.TrimSpace(line)

			pattern := "^\"" + keyword + "\""
			// Get Info interface
			r, _ := regexp.Compile(pattern)
			if r.MatchString(line) {
				r, _ := regexp.Compile("\\s\"(.*?)\"")
				submatch := r.FindStringSubmatch(line)
				value := submatch[1]
				valueList = append(valueList, value)
				isFound = true
				break
			}

		}
		if isFound == false {
			valueList = append(valueList, "")
		} else {
			isFound = false
		}

	}
	return valueList
}

func GetAllDefaultIP() ([]Host, error) {

	// Dummy value ----------------------
	var inventory InventoryInfo
	var inventoryList []InventoryInfo

	inventory.ClientName = "client1"
	inventory.ClientOS = "linux"
	inventoryList = append(inventoryList, inventory)

	inventory.ClientName = "client3"
	inventory.ClientOS = "linux"
	inventoryList = append(inventoryList, inventory)

	inventory.ClientName = "client2"
	inventory.ClientOS = "linux"
	inventoryList = append(inventoryList, inventory)

	// End dummy value --------------------

	var hostList []Host
	var error error

	for _, node := range inventoryList {
		var host Host
		host.ClientName = node.ClientName
		if isConnected := utils.IsNodeReachable(node.ClientName); isConnected == true {
			host.IsConnected = true
			if node.ClientOS == "linux" {
				Output, _ := exec.Command("ansible", node.ClientName, "-m", "setup", "-a", "filter=ansible_default_ipv4").Output()
				ansibleOutput := string(Output)
				valueIPv4 := extractJsonValue(ansibleOutput, []string{"address", "netmask", "gateway", "interface"})

				var ipv4 IPv4Info
				var ipv6 IPv6Info
				var Interfaces Interfaces

				ipv4.Address = valueIPv4[0]
				ipv4.Netmask = valueIPv4[1]
				ipv4.Gateway = valueIPv4[2]
				Interfaces.InterfacesName = valueIPv4[3]
				Interfaces.Ipv4 = ipv4

				Output, _ = exec.Command("ansible", node.ClientName, "-m", "setup", "-a", "filter=ansible_default_ipv6").Output()
				ansibleOutput = string(Output)
				valueIPv6 := extractJsonValue(ansibleOutput, []string{"address", "gateway"})
				ipv6.Address = valueIPv6[0]
				ipv6.Gateway = valueIPv6[1]
				Interfaces.Ipv6 = ipv6

				host.Interfaces = append(host.Interfaces, Interfaces)

			}

		} else {
			host.IsConnected = false
		}
		hostList = append(hostList, host)
	}

	return hostList, error
}
