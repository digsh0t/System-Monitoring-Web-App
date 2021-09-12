package models

import (
	"errors"

	"github.com/wintltr/login-api/database"
)

type PcInfo struct {
	SshConnectionId       int    `json:"id"`
	SshConnectionHostName string `json:"hostnameSSH"`
	State                 string `json:"state"`
}

func GetAllPC() ([]PcInfo, error) {
	db := database.ConnectDB()
	defer db.Close()

	var pcInfoList []PcInfo
	selDB, err := db.Query("SELECT sc_connection_id, sc_hostname FROM ssh_connections")
	if err != nil {
		return pcInfoList, err
	}

	var pcInfo PcInfo
	for selDB.Next() {
		var id int
		var name string

		err = selDB.Scan(&id, &name)
		if err != nil {
			return pcInfoList, err
		}
		pcInfo.SshConnectionId = id
		pcInfo.SshConnectionHostName = name
		pcInfoList = append(pcInfoList, pcInfo)
	}

	return pcInfoList, err

}

func GetAllPcsState() ([]PcInfo, error) {
	var (
		pcsState []PcInfo
		err      error
	)
	pcsState, err = GetAllPC()
	if err != nil {
		return pcsState, errors.New("fail to get all pc")

	}
	for index, pc := range pcsState {
		var state string
		sshConnection, err := GetSSHConnectionFromId(pc.SshConnectionId)
		if err != nil {
			return pcsState, errors.New("fail to get sshConnection")
		}
		result, err := RunCommandFromSSHConnection(*sshConnection, "whoami")
		if err != nil {
			state = "shutdown"
		} else if err == nil && result != "" {
			state = "running"
		}
		pcsState[index].State = state
	}
	return pcsState, err
}
