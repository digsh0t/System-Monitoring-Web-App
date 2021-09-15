package models

import "encoding/json"

type Package struct {
	Name             string `json:"name"`
	Version          string `json:"version"`
	Source           string `json:"source"`
	Size             string `json:"size"`
	Arch             string `json:"arch"`
	PidWithNamespace int    `json:"pid_with_namespace"`
	MountNamespaceId string `json:"mount_namespace_id"`
}

func GetInstalledRPMPackages(sshConnection SshConnectionInfo) ([]Package, error) {
	result, err := sshConnection.RunCommandFromSSHConnectionUseKeys(`osqueryi --json "SELECT * FROM rpm_packages"`)
	if err != nil {
		return nil, err
	}
	var installedRPMs []Package

	err = json.Unmarshal([]byte(result), &installedRPMs)
	return installedRPMs, err
}

func GetInstalledDebPackages(sshConnection SshConnectionInfo) ([]Package, error) {
	result, err := sshConnection.RunCommandFromSSHConnectionUseKeys(`osqueryi --json "SELECT * FROM deb_packages"`)
	if err != nil {
		return nil, err
	}
	var installedDebs []Package

	err = json.Unmarshal([]byte(result), &installedDebs)
	return installedDebs, err
}
