package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/Jeffail/gabs"
)

type osVersion struct {
	Arch         string `json:"arch"`
	Build        string `json:"build"`
	Codename     string `json:"codename"`
	InstallDate  string `json:"install_date"`
	Major        string `json:"major"`
	Minor        string `json:"minor"`
	Name         string `json:"name"`
	Patch        string `json:"patch"`
	Platform     string `json:"platform"`
	PlatformLike string `json:"platform_like"`
	Version      string `json:"version"`
}

func parseOSVersion(output string) (osVersion, error) {
	var osVersionList []osVersion
	err := json.Unmarshal([]byte(output), &osVersionList)
	return osVersionList[0], err
}

func (sshConnection SshConnectionInfo) GetOSVersion() (osVersion, error) {
	var os osVersion
	result, err := sshConnection.RunCommandFromSSHConnectionUseKeys(`osqueryi --json "SELECT * FROM os_version"`)
	if err != nil {
		return os, err
	}
	os, err = parseOSVersion(result)
	return os, err
}

func ParseAnsibleFactsOutput(output string) error {
	openingTag := strings.Index(output, "{")
	if openingTag < 0 {
		return errors.New("Ansible Output is not JSON format")
	}
	closingTag := strings.LastIndex(output, "}")
	if closingTag < 0 {
		return errors.New("Ansible Output is not JSON format")
	}
	jsonStr := output[openingTag:closingTag]
	jsonStr += "}"
	jsonParsed, err := gabs.ParseJSON([]byte(jsonStr))
	if err != nil {
		return err
	}
	fmt.Println(jsonParsed.Path("ansible_facts.architecture").Data())
	fmt.Println(jsonParsed.Path("ansible_facts.windows_domain").Data())
	fmt.Println(jsonParsed.Path("ansible_facts.uptime_seconds").Data())
	fmt.Println(jsonParsed.Path("ansible_facts.hostname").Data())
	fmt.Println(jsonParsed.Path("ansible_facts.memtotal_mb").Data())
	fmt.Println(jsonParsed.Path("ansible_facts.distribution").Data())
	fmt.Println(jsonParsed.Path("ansible_facts.distribution_version").Data())
	fmt.Println(jsonParsed.Path("ansible_facts.bios_date").Data())
	fmt.Println(jsonParsed.Path("ansible_facts.bios_version").Data())
	return nil
}
