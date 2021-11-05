package models

import (
	"database/sql"
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/Jeffail/gabs"
	"github.com/wintltr/login-api/database"
	"github.com/wintltr/login-api/utils"
)

type SshConnectionInformation struct {
	InformationId   int    `json:"informationId"`
	OsName          string `json:"osName"`
	OsVersion       string `json:"osVersion"`
	InstallDate     string `json:"installDate"`
	Serial          string `json:"serial"`
	Hostname        string `json:"hostname"`
	Manufacturer    string `json:"manufacturer"`
	Model           string `json:"model"`
	Architecture    string `json:"architecture"`
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

	// Create json
	type tmpJson struct {
		Host string `json:"host"`
	}

	tmp := tmpJson{
		Host: sshConnection.HostNameSSH,
	}

	tmpJsonMarshal, err := json.Marshal(tmp)
	if err != nil {
		log.Println("fail to unmarshal json")
	}

	// If sshconnection is not network device
	if !sshConnection.IsNetwork {
		// Get os version
		os, err := sshConnection.GetOSVersion()
		if err != nil {
			log.Println("fail to get os version")
		} else {
			// Get Name
			sshConnectionInformation.OsName = os.Name

			// Get Version
			sshConnectionInformation.OsVersion = os.Version
		}

		// Get install date
		sshConnectionInformation.InstallDate, err = GetInstallDate(sshConnection)
		if err != nil {
			log.Println("fail to get install date")
		}

		// Get Serial
		sshConnectionInformation.Serial, err = GetClientSerial(sshConnection)
		if err != nil {
			log.Println("fail to get serial")
		}

		// Get Hostname
		sshConnectionInformation.Hostname, err = GetHostname(sshConnection)
		if err != nil {
			log.Println("fail to get hostname")
		}

		var filepath = "./yamls/get_client_setup.yml"
		output, err := RunAnsiblePlaybookWithjson(filepath, string(tmpJsonMarshal))
		if err != nil {
			log.Println("fail to run playbook")
		}
		// Get substring from ansible output
		data := utils.ExtractSubString(output, " => ", "PLAY RECAP")

		// Parse Json format
		jsonParsed, err := gabs.ParseJSON([]byte(data))
		if err != nil {
			log.Println("fail to parse json")
		}

		// Get Product Name
		sshConnectionInformation.Model = strings.Trim(strings.TrimSpace(jsonParsed.Search("msg", "ansible_facts", "ansible_product_name").String()), "\"")

		// Get Manufacturer
		sshConnectionInformation.Manufacturer = strings.Trim(strings.TrimSpace(jsonParsed.Search("msg", "ansible_facts", "ansible_system_vendor").String()), "\"")

		// Get Architecture
		sshConnectionInformation.Architecture = strings.Trim(strings.TrimSpace(jsonParsed.Search("msg", "ansible_facts", "ansible_architecture").String()), "\"")

	} else {
		// Network device
		var filepath string
		if sshConnection.NetworkOS == "ios" {
			filepath = "./yamls/network_client/cisco/cisco_getfacts.yml"
		} else if sshConnection.NetworkOS == "vyos" {
			filepath = "./yamls/network_client/vyos/vyos_getfacts.yml"
		} else if sshConnection.NetworkOS == "junos" {
			filepath = "./yamls/network_client/juniper/juniper_getfacts.yml"
		}
		output, err := RunAnsiblePlaybookWithjson(filepath, string(tmpJsonMarshal))
		if err != nil {
			log.Println("fail to run playbook")
		}
		// Get substring from ansible output
		data := utils.ExtractSubString(output, " => ", "PLAY RECAP")

		// Parse Json format
		jsonParsed, err := gabs.ParseJSON([]byte(data))
		if err != nil {
			log.Println("fail to parse json")
		}

		// Serial num
		rawSerial := strings.Trim(strings.TrimSpace(jsonParsed.Search("msg", "ansible_facts", "ansible_net_serialnum").String()), "\"")
		if rawSerial == "null" {
			sshConnectionInformation.Serial = ""
		} else {
			sshConnectionInformation.Serial = rawSerial
		}

		// Get os name
		sshConnectionInformation.OsName = strings.Trim(strings.TrimSpace(jsonParsed.Search("msg", "ansible_facts", "ansible_net_system").String()), "\"")

		// Get os version
		sshConnectionInformation.OsVersion = strings.Trim(strings.TrimSpace(jsonParsed.Search("msg", "ansible_facts", "ansible_net_version").String()), "\"")
		if sshConnection.NetworkOS == "vyos" {
			sshConnectionInformation.OsVersion = strings.TrimRight(sshConnectionInformation.OsVersion, "m[b10u\\")
		}

		// Get Hostname
		sshConnectionInformation.Hostname = strings.Trim(strings.TrimSpace(jsonParsed.Search("msg", "ansible_facts", "ansible_net_hostname").String()), "\"")
		if sshConnection.NetworkOS == "vyos" {
			sshConnectionInformation.Hostname = strings.TrimRight(sshConnectionInformation.Hostname, "m[b10u\\")
		}

	}

	// Insert to DB
	_, err = sshConnectionInformation.AddSSHConnectionInformationToDB()
	if err != nil {
		log.Println("fail to insert sshConnection Information to DB")
	}

	return result, err
}

func GetHostname(sshConnection SshConnectionInfo) (string, error) {
	var (
		hostname string
		err      error
	)
	output, err := sshConnection.RunCommandFromSSHConnectionUseKeys(`osqueryi --json "SELECT hostname FROM system_info"`)
	if err != nil {
		return hostname, err
	}

	type HostnameStruct struct {
		Hostname string `json:"hostname"`
	}
	var hostnameList []HostnameStruct
	err = json.Unmarshal([]byte(output), &hostnameList)
	if err != nil {
		return hostname, err
	}
	hostname = hostnameList[0].Hostname
	return hostname, err
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

	query = "INSERT INTO ssh_connections_information (sc_info_osname, sc_info_osversion, sc_info_installdate, sc_info_serial, sc_info_hostname, sc_info_manufacturer, sc_info_model, sc_info_architecture, sc_info_connection_id) VALUES (?,?,?,?,?,?,?,?,?)"
	stmt, err := db.Prepare(query)
	if err != nil {
		return lastId, err
	}
	defer stmt.Close()

	var res sql.Result

	res, err = stmt.Exec(sshConnectionInformation.OsName, sshConnectionInformation.OsVersion, sshConnectionInformation.InstallDate, sshConnectionInformation.Serial, sshConnectionInformation.Hostname, sshConnectionInformation.Manufacturer, sshConnectionInformation.Model, sshConnectionInformation.Architecture, sshConnectionInformation.SshConnectionId)

	if err != nil {
		return lastId, err
	}
	lastId, err = res.LastInsertId()
	if err != nil {
		return lastId, err
	}

	return lastId, err
}

func GetDetailOSReport(osType string) ([]SshConnectionInformation, error) {
	var (
		sshConnectionInfoList []SshConnectionInformation
		err                   error
	)
	db := database.ConnectDB()
	defer db.Close()
	var query string
	if osType == "" {
		query = `SELECT sc_info_id, sc_info_osname, sc_info_osversion, sc_info_installdate, sc_info_serial, sc_info_hostname, sc_info_manufacturer, sc_info_model, sc_info_architecture, sc_info_connection_id FROM ssh_connections_information `
	} else {
		if osType == "Linux" {
			query = `SELECT sc_info_id, sc_info_osname, sc_info_osversion, sc_info_installdate, sc_info_serial, sc_info_hostname, sc_info_manufacturer, sc_info_model, sc_info_architecture, sc_info_connection_id FROM ssh_connections_information WHERE sc_info_osname='Ubuntu' or sc_info_osname LIKE '%CentOS%' or sc_info_osname LIKE '%Kali%'`
		} else if osType == "Windows" {
			query = `SELECT sc_info_id, sc_info_osname, sc_info_osversion, sc_info_installdate, sc_info_serial, sc_info_hostname, sc_info_manufacturer, sc_info_model, sc_info_architecture, sc_info_connection_id FROM ssh_connections_information WHERE sc_info_osname LIKE '%Windows%'`
		} else if osType == "Network" {
			query = `SELECT sc_info_id, sc_info_osname, sc_info_osversion, sc_info_installdate, sc_info_serial, sc_info_hostname, sc_info_manufacturer, sc_info_model, sc_info_architecture, sc_info_connection_id FROM ssh_connections_information WHERE sc_info_osname='ios' or sc_info_osname='junos' or sc_info_osname='vyos'`
		} else {
			return sshConnectionInfoList, err
		}
	}
	selDB, err := db.Query(query)
	if err != nil {
		return nil, err
	}

	var sshConnectionInfo SshConnectionInformation
	for selDB.Next() {
		err = selDB.Scan(&sshConnectionInfo.InformationId, &sshConnectionInfo.OsName, &sshConnectionInfo.OsVersion, &sshConnectionInfo.InstallDate, &sshConnectionInfo.Serial, &sshConnectionInfo.Hostname, &sshConnectionInfo.Manufacturer, &sshConnectionInfo.Model, &sshConnectionInfo.Architecture, &sshConnectionInfo.SshConnectionId)
		if err != nil {
			return nil, err
		}
		sshConnectionInfoList = append(sshConnectionInfoList, sshConnectionInfo)
	}
	return sshConnectionInfoList, err

}

func GetSSHConnectionInformationBySSH_Id(sshConnectionId int) (SshConnectionInformation, error) {
	var (
		sshConnectionInfo SshConnectionInformation
		err               error
	)
	db := database.ConnectDB()
	defer db.Close()

	query := `SELECT sc_info_id, sc_info_osname, sc_info_osversion, sc_info_installdate, sc_info_serial, sc_info_hostname, sc_info_manufacturer, sc_info_model, sc_info_architecture, sc_info_connection_id FROM ssh_connections_information WHERE sc_info_connection_id = ? `

	selDB, err := db.Query(query, sshConnectionId)
	if err != nil {
		return sshConnectionInfo, err
	}

	for selDB.Next() {
		err = selDB.Scan(&sshConnectionInfo.InformationId, &sshConnectionInfo.OsName, &sshConnectionInfo.OsVersion, &sshConnectionInfo.InstallDate, &sshConnectionInfo.Serial, &sshConnectionInfo.Hostname, &sshConnectionInfo.Manufacturer, &sshConnectionInfo.Model, &sshConnectionInfo.Architecture, &sshConnectionInfo.SshConnectionId)
		if err != nil {
			return sshConnectionInfo, err
		}
	}
	return sshConnectionInfo, err
}

func (sshConnection SshConnectionInfo) GetDetailSSHConInfo() (SshConnectionInformation, error) {
	var information SshConnectionInformation
	db := database.ConnectDB()
	defer db.Close()
	query := `SELECT sc_info_id, sc_info_osname, sc_info_osversion, sc_info_installdate, sc_info_serial, sc_info_connection_id FROM ssh_connections_information WHERE sc_info_connection_id = ?`
	row := db.QueryRow(query, sshConnection.SSHConnectionId)
	err := row.Scan(&information.InformationId, &information.OsName, &information.OsVersion, &information.InstallDate, &information.Serial, &information.SshConnectionId)
	return information, err
}
