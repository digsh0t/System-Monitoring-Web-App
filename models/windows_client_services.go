package models

import "encoding/json"

type services struct {
	Description     string `json:"description"`
	DisplayName     string `json:"display_name"`
	ModulePath      string `json:"module_path"`
	Name            string `json:"name"`
	Path            string `json:"path"`
	Pid             string `json:"pid"`
	ServiceExitCode string `json:"service_exit_code"`
	ServiceType     string `json:"service_type"`
	StartType       string `json:"start_type"`
	Status          string `json:"status"`
	UserAccount     string `json:"user_account"`
	Win32ExitCode   string `json:"win32_exit_code"`
}

func parseServiceList(output string) ([]services, error) {
	var serviceList []services
	err := json.Unmarshal([]byte(output), &serviceList)
	return serviceList, err
}

func (sshConnection SshConnectionInfo) GetServiceList() ([]services, error) {
	var serviceList []services
	result, err := sshConnection.RunCommandFromSSHConnectionUseKeys(`osqueryi --json "SELECT * FROM services`)
	if err != nil {
		return serviceList, err
	}
	serviceList, err = parseServiceList(result)
	return serviceList, err
}

func (sshConnection *SshConnectionInfo) ChangeWindowsServiceState(serviceName string, serviceState string) error {
	type changedServiceState struct {
		Host         string `json:"host"`
		ServiceName  string `json:"service_name"`
		ServiceState string `json:"service_state"`
	}
	marshalled, err := json.Marshal(changedServiceState{Host: sshConnection.HostNameSSH, ServiceName: serviceName, ServiceState: serviceState})
	if err != nil {
		return err
	}
	RunAnsiblePlaybookWithjson("./yamls/windows_client/modify_windows_service_state.yml", string(marshalled))
	return err
}
