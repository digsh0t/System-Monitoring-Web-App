package models

import (
	"errors"

	"github.com/wintltr/login-api/database"
)

func (sshConnection *SshConnectionInfo) AddNewWatcher(watchList string) error {
	db := database.ConnectDB()
	defer db.Close()

	query := "INSERT INTO ssh_connection_alert (sca_id, sca_connection_name, sca_alert_pri) VALUES (?,?,?)"
	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(sshConnection.SSHConnectionId, sshConnection.HostNameSSH, watchList)
	return err
}

func RemoveWatcher(sshConnectionId int) error {
	db := database.ConnectDB()
	defer db.Close()

	stmt, err := db.Prepare("DELETE FROM ssh_connection_alert WHERE sca_id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	res, err := stmt.Exec(sshConnectionId)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if rows == 0 {
		return errors.New("This SSH Connection is not in watchlist")
	}
	return err
}
