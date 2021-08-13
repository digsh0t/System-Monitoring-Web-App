package models

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

type DefaultIPInfo struct {
	ClientName string `json:"clientName"`
	Interfaces string `json:"interface"`
	Address    string `json:"address"`
	Netmask    string `json:"netmask"`
	Gateway    string `json:"gateway"`
}

func (defaultIP *DefaultIPInfo) GetAllDefaultIP() ([]DefaultIPInfo, error) {
	module := "ansible"
	arg1 := "all"
	arg2 := "-m"
	arg3 := "setup"
	arg4 := "-a"
	arg5 := "filter=ansible_default_ipv4"
	output, _ := exec.Command(module, arg1, arg2, arg3, arg4, arg5).Output()
	outputString := string(output)

	var defaultIPList []DefaultIPInfo
	var error error

	for _, line := range strings.Split(strings.TrimRight(outputString, "\n"), "\n") {
		line = strings.TrimSpace(line)
		fmt.Println(line)

		// Get clientName
		r, _ := regexp.Compile("SUCCESS")
		if r.MatchString(line) {
			sizeArrange := r.FindStringIndex(line)
			firstIndex := sizeArrange[0] - 3
			clientName := line[:firstIndex]
			defaultIP.ClientName = clientName
			fmt.Println("name:", clientName)
		}

		// Get Address IP
		r, _ = regexp.Compile("^\"address\"")
		if r.MatchString(line) {
			r, _ = regexp.Compile("[0-9]+.[0-9]+.[0-9]+.[0-9]+")
			address := r.FindString(line)
			defaultIP.Address = address
			fmt.Println("add:", address)

		}

		// Get Subnet Mask
		r, _ = regexp.Compile("^\"netmask\"")
		if r.MatchString(line) {
			r, _ = regexp.Compile("[0-9]+.[0-9]+.[0-9]+.[0-9]+")
			netmask := r.FindString(line)
			defaultIP.Netmask = netmask
			fmt.Println("subnet:", netmask)

		}

		// Get Default Gateway
		r, _ = regexp.Compile("^\"gateway\"")
		if r.MatchString(line) {
			r, _ = regexp.Compile("[0-9]+.[0-9]+.[0-9]+.[0-9]+")
			gateway := r.FindString(line)
			defaultIP.Gateway = gateway
			fmt.Println("gateway:", gateway)

		}

		// Get Default interface
		r, _ = regexp.Compile("^\"interface\"")
		if r.MatchString(line) {
			r, _ := regexp.Compile("\\s\"([a-z0-9]+)\"")
			arrayString := r.FindStringSubmatch(line)
			defaultIP.Interfaces = arrayString[1]
			fmt.Println("faces:", defaultIP.Interfaces)
		}

		// Fetch object into List and Clear data.
		if defaultIP.Netmask != "" {
			defaultIPList = append(defaultIPList, *defaultIP)
			defaultIP.ClientName = ""
			defaultIP.Netmask = ""
			defaultIP.Address = ""
			defaultIP.Gateway = ""
			defaultIP.Interfaces = ""

		}

	}

	return defaultIPList, error
}
