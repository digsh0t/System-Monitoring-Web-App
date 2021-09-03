package models

import (
	"errors"
	"time"

	"github.com/wintltr/login-api/database"
)

type SysInfo struct {
	ConnectionId int       `json:"id"`
	HostnameSSH  string    `json:"hostnameSSH"`
	AvgCPU       string    `json:"avgcpu"`
	AvgMem       string    `json:"avgmem"`
	Timestamp    string    `json:"timestamp"`
	UfwStatus    bool      `json:"ufwstatus"`
	UfwRules     []UfwRule `json:"ufwrulelist"`
}

type OnlineStatus struct {
	ConnectionId int
	IsOn         bool
}

func InsertSysInfoToDB(sysInfo SysInfo, ip string, hostname string, connectionId int) error {
	db := database.ConnectDB()
	defer db.Close()

	stmt, err := db.Prepare("INSERT INTO sys_info_logs (syl_hostname, syl_avg_cpu, syl_avg_mem, syl_timestamp, syl_connection_id) VALUES (?,?,?,?,?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(hostname, sysInfo.AvgCPU, sysInfo.AvgMem, sysInfo.Timestamp, connectionId)
	if err != nil {
		return err
	}
	return err
}

func GetLatestSysInfo(sshConnectionId int, interval int) (SysInfo, error) {
	db := database.ConnectDB()
	defer db.Close()

	var sysInfo SysInfo
	row := db.QueryRow("SELECT syl_hostname, syl_avg_cpu, syl_avg_mem, syl_timestamp, syl_connection_id FROM sys_info_logs WHERE syl_connection_id = ? ORDER BY syl_id DESC LIMIT 1", sshConnectionId)
	err := row.Scan(&sysInfo.HostnameSSH, &sysInfo.AvgCPU, &sysInfo.AvgMem, &sysInfo.Timestamp, &sysInfo.ConnectionId)

	if err != nil && err.Error() != "sql: no rows in result set" {
		return sysInfo, errors.New("fail to retrieve ssh connection info")
	}
	layout := "01-02-2006 15:04:05"
	t, _ := time.Parse(layout, sysInfo.Timestamp)

	current, _ := time.Parse(layout, time.Now().Format("01-02-2006 15:04:05"))
	diff := current.Sub(t)
	sshConnection, _ := GetSSHConnectionFromId(sshConnectionId)
	if diff.Seconds() > float64(interval) {
		sysInfo = SysInfo{}
		sysInfo.ConnectionId = sshConnectionId
		sysInfo.HostnameSSH = sshConnection.HostNameSSH
	}
	return sysInfo, err
}

func GetAllSysInfo(sshConnectionList []SshConnectionInfo) ([]SysInfo, error) {
	var sysInfoList []SysInfo
	var sysInfo SysInfo
	var err error

	for _, sshConnection := range sshConnectionList {
		sysInfo, err = GetLatestSysInfo(sshConnection.SSHConnectionId, 10)
		if err != nil {
			return sysInfoList, err
		}
		sysInfoList = append(sysInfoList, sysInfo)
	}

	return sysInfoList, nil
}

func CheckOnlineStatus(sshConnectionlist []SshConnectionInfo) []OnlineStatus {
	var statuses []OnlineStatus
	for _, sshConnection := range sshConnectionlist {
		sysinfo, _ := GetLatestSysInfo(sshConnection.SSHConnectionId, 100)
		if sysinfo.AvgCPU == "" {
			statuses = append(statuses, OnlineStatus{sshConnection.SSHConnectionId, false})
		} else {
			statuses = append(statuses, OnlineStatus{sshConnection.SSHConnectionId, true})
		}
	}
	return statuses
}
