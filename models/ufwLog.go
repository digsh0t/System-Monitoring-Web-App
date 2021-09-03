package models

import (
	"encoding/json"
	"errors"

	"github.com/wintltr/login-api/database"
)

type UfwRule struct {
	To     string `json:"to"`
	Action string `json:"action"`
	From   string `json:"from"`
}

type UfwList struct {
	SshConnectionId int       `json:"sshconnectionid"`
	UfwHostname     string    `json:"ufwHostname"`
	UfwId           int       `json:"ufwid"`
	UfwStatus       bool      `json:"ufwstatus"`
	UfwRules        []UfwRule `json:"ufwrules"`
}

func GetUfwRulesFromDB(sshConnectionId int) (UfwList, error) {
	db := database.ConnectDB()
	defer db.Close()

	var ufwList UfwList
	var tmp string

	row := db.QueryRow("SELECT cus_id, cus_hostname, cus_ufw_status, cus_ufw_rules, cus_connection_id FROM client_ufw_settings WHERE cus_id = ?", sshConnectionId)
	err := row.Scan(&ufwList.UfwId, &ufwList.UfwHostname, &ufwList.UfwStatus, &tmp, &ufwList.SshConnectionId)
	if row == nil {
		return ufwList, errors.New("ufw with ssh connection doesn't exist")
	}
	if err != nil {
		return ufwList, errors.New("fail to retrieve ufw info")
	}
	err = json.Unmarshal([]byte(tmp), &ufwList.UfwRules)
	if err != nil {
		return ufwList, errors.New("fail to get ufw info from database")
	}

	return ufwList, err
}

func InsertUfwToDB(hostname string, sshConnectionId int, ufwStatus bool, ufwRules []UfwRule) error {
	db := database.ConnectDB()
	defer db.Close()

	blobed, err := json.Marshal(ufwRules)
	if err != nil {
		return err
	}

	stmt, err := db.Prepare("INSERT INTO client_ufw_settings (cus_id, cus_hostname, cus_ufw_status, cus_ufw_rules, cus_connection_id) VALUES (?,?,?,?,?) ON DUPLICATE KEY UPDATE cus_ufw_rules = ?, cus_ufw_status = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(sshConnectionId, hostname, ufwStatus, blobed, sshConnectionId, blobed, ufwStatus)
	if err != nil {
		return err
	}
	return err
}

// func UpdateUfwRuleToDB(ufwList []UfwRule) error {
// 	for i, ufwRule := range ufwList {

// 	}
// }
