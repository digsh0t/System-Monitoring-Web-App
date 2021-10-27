package models

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/wintltr/login-api/database"
)

type SshConnectionInformation struct {
	InformationId   int    `json:"informationId"`
	OsName          string `json:"osName"`
	OsVersion       string `json:"osVersion"`
	InstallDate     string `json:"installDate"`
	Signature       string `json:"signature"`
	SshConnectionId int    `json:"sshConnectionId"`
}

func AddSSHConnectionInformation(sshConnection SshConnectionInfo, lastId int64) (bool, error) {
	var (
		sshConnectionInformation SshConnectionInformation
		result                   bool
		err                      error
	)

	// Get SshConnectionId
	sshConnectionInformation.SshConnectionId = int(lastId)
	if !sshConnection.IsNetwork {

		// Get os version
		os, err := sshConnection.GetOSVersion()
		if err != nil {
			return false, errors.New("fail to get os version")
		}
		// Get Name
		sshConnectionInformation.OsName = os.Name

		// Get Version
		sshConnectionInformation.OsVersion = os.Version

		// Get install date
		sshConnectionInformation.InstallDate, err = GetInstallDate(sshConnection)
		if err != nil {
			return false, errors.New("fail to get install date")
		}

		// Get Signature

	} else {
		// not supported for network device
	}

	// Insert to DB
	_, err = sshConnectionInformation.AddSSHConnectionInformationToDB()
	if err != nil {
		fmt.Println(err.Error())
		return false, errors.New("fail to insert sshConnection Information to DB")
	}

	return result, err
}

func GetInstallDate(sshConnection SshConnectionInfo) (string, error) {
	var (
		installDate string
		err         error
	)
	if strings.Contains(sshConnection.OsType, "Windows") {
		output, err := sshConnection.RunCommandFromSSHConnectionUseKeys("systeminfo|find /i \"original\"")
		if err != nil {
			return installDate, err
		}
		// Get Time
		tmp := strings.Split(output, "    ")

		// Convert Time format
		rawInstalledDate := strings.TrimRight(strings.TrimSpace(tmp[1]), "MP")
		rawInstalledDate = strings.ReplaceAll(rawInstalledDate, ",", "")
		dt, _ := time.Parse("1/2/2006 3:4:5", strings.TrimSpace(rawInstalledDate))
		installDate = dt.Format("2006-01-02 3:4:5")

	} else if strings.Contains(sshConnection.OsType, "Ubuntu") || strings.Contains(sshConnection.OsType, "CentOS") || strings.Contains(sshConnection.OsType, "Kali") {
		installDate, err = sshConnection.RunCommandFromSSHConnectionUseKeys("ls -lact --full-time /etc | tail -1 | awk '{print $6,$7}'")
		if err != nil {
			return installDate, err
		}
	}
	return installDate, err
}

func (sshConnectionInformation *SshConnectionInformation) AddSSHConnectionInformationToDB() (int64, error) {
	db := database.ConnectDB()
	defer db.Close()

	var query string
	var lastId int64

	// Use key-base Authentication

	query = "INSERT INTO ssh_connections_information (sc_info_osname, sc_info_osversion, sc_info_installdate, sc_info_signature, sc_info_connection_id) VALUES (?,?,?,?,?)"
	stmt, err := db.Prepare(query)
	if err != nil {
		return lastId, err
	}
	defer stmt.Close()

	var res sql.Result

	res, err = stmt.Exec(sshConnectionInformation.OsName, sshConnectionInformation.OsVersion, sshConnectionInformation.InstallDate, sshConnectionInformation.Signature, sshConnectionInformation.SshConnectionId)

	if err != nil {
		return lastId, err
	}
	lastId, err = res.LastInsertId()
	if err != nil {
		return lastId, err
	}

	return lastId, err
}
