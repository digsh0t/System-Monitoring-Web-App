package models

import (
	"errors"

	"github.com/wintltr/login-api/database"
)

type SNMPInfo struct {
	SnmpID          int    `json:"snmpID"`
	AuthUsername    string `json:"authUsername"`
	AuthPassword    string `json:"authPassword"`
	PrivPassword    string `json:"privPassword"`
	SSHConnectionID int    `json:"sshConnectionId"`
}

func (snmp *SNMPInfo) AddSNMPConnectionToDB() (bool, error) {
	db := database.ConnectDB()
	defer db.Close()

	query := "INSERT INTO snmp_credential (snmp_auth_username, snmp_auth_password, snmp_priv_password, snmp_connection_id) VALUES (?,?,?,?)"
	stmt, err := db.Prepare(query)
	if err != nil {
		return false, err
	}
	defer stmt.Close()

	authPasswordEncrypted := AESEncryptKey(snmp.AuthPassword)
	privPasswordEncrypted := AESEncryptKey(snmp.PrivPassword)
	_, err = stmt.Exec(snmp.AuthUsername, authPasswordEncrypted, privPasswordEncrypted, snmp.SSHConnectionID)
	if err != nil {
		return false, err
	}

	return true, err
}

func GetSNMPCredentialFromSshConnectionId(sshConnectionId int) (SNMPInfo, error) {
	db := database.ConnectDB()
	defer db.Close()

	var snmpCredential SNMPInfo
	row := db.QueryRow("SELECT snmp_id, snmp_auth_username, snmp_auth_password, snmp_priv_password, snmp_connection_id FROM snmp_credential WHERE snmp_connection_id = ?", sshConnectionId)
	var (
		encryptedAuthPassword string
		encryptedPrivPassword string
	)
	err := row.Scan(&snmpCredential.SnmpID, &snmpCredential.AuthUsername, &encryptedAuthPassword, &encryptedPrivPassword, &snmpCredential.SSHConnectionID)
	if row == nil {
		return snmpCredential, errors.New("ssh connection doesn't exist")
	}
	if err != nil {
		return snmpCredential, errors.New("fail to retrieve ssh connection info")
	}

	// Decrypted Password
	snmpCredential.AuthPassword, err = AESDecryptKey(encryptedAuthPassword)
	if err != nil {
		return snmpCredential, errors.New("fail to decrypt password")
	}
	snmpCredential.PrivPassword, err = AESDecryptKey(encryptedPrivPassword)
	if err != nil {
		return snmpCredential, errors.New("fail to decrypt password")
	}

	return snmpCredential, err
}
