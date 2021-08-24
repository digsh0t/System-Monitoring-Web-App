package models

import (
	"fmt"

	"github.com/wintltr/login-api/database"
)

type UfwRule struct {
	To     string `json:"to"`
	Action string `json:"action"`
	From   string `json:"from"`
}

func InsertUfwToDB(hostname string, sshConnectionId int, ufwStatus bool, ufwRules []UfwRule) error {
	db := database.ConnectDB()
	defer db.Close()

	blobed := []byte(fmt.Sprintf("%v", ufwRules))

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
