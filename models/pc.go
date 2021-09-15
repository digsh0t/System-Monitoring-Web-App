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

func GetPcStateByID(sshConnectionId int) (string, error) {
	var pcState string

	sshConnection, err := GetSSHConnectionFromId(sshConnectionId)
	if err != nil {
		return pcState, errors.New("fail to get sshConnection")
	}

	// Run remote command to check pc "running" or "shutdown"
	result, err := sshConnection.RunCommandFromSSHConnectionUseKeys("whoami")
	if err != nil {
		pcState = "shutdown"
		// Avoid returning error make function working not correctly
		err = nil
	} else if err == nil && result != "" {
		pcState = "running"
	}

	return pcState, err
}
