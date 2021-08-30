package models

import "github.com/wintltr/login-api/database"

type PcInfo struct {
	SshConnectionId       int    `json:"id"`
	SshConnectionHostName string `json:"hostnameSSH"`
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
