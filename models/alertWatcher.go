package models

import "github.com/wintltr/login-api/database"

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
