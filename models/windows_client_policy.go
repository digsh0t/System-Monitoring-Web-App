package models

import (
	"encoding/json"
	"strings"
)

type uacSetting struct {
	Data string `json:"data"`
	Path string `json:"path"`
}

func parseUAC(output string) ([]uacSetting, error) {
	var uacList []uacSetting
	err := json.Unmarshal([]byte(output), &uacList)
	return uacList, err
}

func (sshConnection SshConnectionInfo) GetUACSettings() ([]uacSetting, error) {
	var uacList []uacSetting
	result, err := sshConnection.RunCommandFromSSHConnectionUseKeys(`osqueryi --json "SELECT data, path FROM registry WHERE key = 'HKEY_LOCAL_MACHINE\SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\Policies\System'"`)
	if err != nil {
		return uacList, err
	}
	uacList, err = parseUAC(result)
	return uacList, err
}

func beautifyUACList(uacList []uacSetting) {

	consentPromptBehaviorUser := map[string]string{
		"0": "Automatically deny elevation requests.",
		"1": "Prompt for credentials on the secure desktop.",
		"3": "Prompt for credentials.",
	}

	consentPromptBehaviorAdmin := map[string]string{
		"0": "Elevate without prompting.",
		"1": "Prompt for credentials on the secure desktop.",
		"2": "Prompt for consent on the secure desktop.",
		"3": "Prompt for credentials.",
		"4": "Prompt for consent.",
		"5": "Prompt for consent for non-Windows binaries.",
	}

	dscAutomationHostEnabled := map[string]string{
		"0": "Disable configuring the machine at boot-up.",
		"1": "Enable configuring the machine at boot-up.",
		"2": "Enable configuring the machine only if DSC is in pending or current state. This is the default value.",
	}

	// uacPath := map[string]string{
	// 	"FilterAdministratorToken":   "Admin Approval Mode for the built-in Administrator account.",
	// 	"EnableUIADesktopToggle":     "Allow UIAccess applications to prompt for elevation without using the secure desktop.",
	// 	"ConsentPromptBehaviorAdmin": "Behavior of the elevation prompt for administrators in Admin Approval Mode.",
	// 	"ConsentPromptBehaviorUser":  "Behavior of the elevation prompt for standard users.",
	// 	"EnableInstallerDetection":   "Detect application installations and prompt for elevation.",
	// 	"ValidateAdminCodeSignatures	": "Only elevate executables that are signed and validated.",
	// 	"EnableSecureUIAPaths":  "Only elevate UIAccess applications that are installed in secure locations.",
	// 	"EnableLUA":             "Run all administrators in Admin Approval Mode.",
	// 	"PromptOnSecureDesktop": "Switch to the secure desktop when prompting for elevation.",
	// 	"EnableVirtualization":  "Virtualize file and registry write failures to per-user locations.",
	// }

	for _, uacSetting := range uacList {
		if strings.Contains(uacSetting.Path, "ConsentPromptBehaviorUser") {
			uacSetting.Data = consentPromptBehaviorUser[uacSetting.Data]
			uacSetting.Path = "Behavior of the elevation prompt for standard users"
		}
		if strings.Contains(uacSetting.Path, "ConsentPromptBehaviorAdmin") {
			uacSetting.Data = consentPromptBehaviorAdmin[uacSetting.Data]
			uacSetting.Path = "Behavior of the elevation prompt for administrators in Admin Approval Mode"
		}
		if strings.Contains(uacSetting.Path, "DSCAutomationHostEnabled") {
			uacSetting.Data = dscAutomationHostEnabled[uacSetting.Data]
			uacSetting.Path = "Behavior of the elevation prompt for administrators in Admin Approval Mode"
		}
	}
}
