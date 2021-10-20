package models

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/wintltr/login-api/database"
	"github.com/wintltr/login-api/utils"
	"golang.org/x/crypto/ssh"
)

type SshConnectionInfo struct {
	SSHConnectionId int    `json:"sshConnectionId"`
	UserSSH         string `json:"userSSH"`
	PasswordSSH     string `json:"passwordSSH"`
	HostNameSSH     string `json:"hostnameSSH"`
	HostSSH         string `json:"hostSSH"`
	PortSSH         int    `json:"portSSH"`
	CreatorId       int    `json:"creatorId"`
	SSHKeyId        int    `json:"sshKeyId"`
	OsType          string `json:"osType"`
	IsNetwork       bool   `json:"isNetwork"`
	NetworkType     string `json:"networkType"`
	NetworkOS       string `json:"networkOS"`
}

//Test SSH connection using username and password
func (sshConnection *SshConnectionInfo) TestConnectionPassword() (bool, error) {
	sshConfig := &ssh.ClientConfig{
		User: sshConnection.UserSSH,
		Auth: []ssh.AuthMethod{
			ssh.Password(sshConnection.PasswordSSH),
		},
		Timeout:         30 * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	sshConfig.Config.KeyExchanges = append(sshConfig.Config.KeyExchanges, "diffie-hellman-group1-sha1", "ecdh-sha2-nistp384")
	cipherOrder := sshConfig.Ciphers
	sshConfig.Ciphers = append(cipherOrder, "aes128-ctr", "aes192-ctr", "aes256-ctr", "arcfour256", "arcfour128", "arcfour", "aes128-cbc")

	addr := fmt.Sprintf("%s:%d", sshConnection.HostSSH, sshConnection.PortSSH)

	_, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		fmt.Println(err.Error())
		return false, err
	} else {
		return true, err
	}
}

//Read private key from private key file
func ProcessPrivateKey(keyId int) (ssh.AuthMethod, error) {
	//buffer, err := ioutil.ReadFile(file)
	//fmt.Println(string(buffer))
	privateKey, _ := GetSSHKeyFromId(keyId)
	decrytedPrivateKey, err := AESDecryptKey(privateKey.PrivateKey)
	if err != nil {
		return nil, err
	}

	buffer := []byte(decrytedPrivateKey)
	// if err != nil {
	// 	return nil, err
	// }

	key, err := ssh.ParsePrivateKey(buffer)
	//If error means Private key is protected by passphrase
	if err != nil {
		utils.EnvInit()
		key, err = ssh.ParsePrivateKeyWithPassphrase(buffer, []byte(os.Getenv("SECRET_SSH_PASSPHRASE")))
		if err != nil {
			return nil, err
		}
		return ssh.PublicKeys(key), err
	}
	return ssh.PublicKeys(key), err
}

//Test the SSH connection using private key
func (sshConnection *SshConnectionInfo) TestConnectionPublicKey() (bool, error) {
	//If private key is incorrect or wrong format, return error immediately
	var auth []ssh.AuthMethod
	authMethod, err := ProcessPrivateKey(sshConnection.SSHKeyId)
	if err != nil {
		return false, err
	}
	//Else continue testing connection using the above key
	auth = append(auth, authMethod)

	sshConfig := &ssh.ClientConfig{
		User:            sshConnection.UserSSH,
		Auth:            auth,
		Timeout:         30 * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// Specific cipher alogithm for connecting network device
	cipherOrder := sshConfig.Ciphers
	sshConfig.Ciphers = append(cipherOrder, "aes128-ctr", "aes192-ctr", "aes256-ctr", "arcfour256", "arcfour128", "arcfour", "aes128-cbc")
	addr := fmt.Sprintf("%s:%d", sshConnection.HostSSH, sshConnection.PortSSH)

	_, err = ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return false, err
	} else {
		return true, err
	}
}

func (sshConnection SshConnectionInfo) CheckSSHConnectionExist() (bool, error) {
	sshConnectionList, err := GetAllSSHConnection()
	if err != nil {
		return true, err
	}
	for _, curConnection := range sshConnectionList {
		if curConnection.HostSSH == sshConnection.HostSSH && curConnection.PortSSH == sshConnection.PortSSH {
			return true, errors.New("This SSH Connection is already exists")
		}
	}
	return false, nil
}

func (sshConnection *SshConnectionInfo) AddSSHConnectionToDB() (int64, error) {
	db := database.ConnectDB()
	defer db.Close()

	var query string
	var lastId int64

	// Use key-base Authentication
	if sshConnection.PasswordSSH == "" {
		query = "INSERT INTO ssh_connections (sc_username, sc_host, sc_hostname, sc_port, creator_id, ssh_key_id, sc_ostype, sc_isnetwork, sc_networktype, sc_networkos) VALUES (?,?,?,?,?,?,?,?,?,?)"
	} else {
		query = "INSERT INTO ssh_connections (sc_username, sc_host, sc_hostname, sc_port, creator_id, sc_password, sc_ostype, sc_isnetwork, sc_networktype, sc_networkos) VALUES (?,?,?,?,?,?,?,?,?,?)"
	}
	stmt, err := db.Prepare(query)
	if err != nil {
		return lastId, err
	}
	defer stmt.Close()

	var res sql.Result
	if sshConnection.PasswordSSH == "" {
		res, err = stmt.Exec(sshConnection.UserSSH, sshConnection.HostSSH, sshConnection.HostNameSSH, sshConnection.PortSSH, sshConnection.CreatorId, sshConnection.SSHKeyId, sshConnection.OsType, sshConnection.IsNetwork, sshConnection.NetworkType, sshConnection.NetworkOS)
	} else {
		encryptedPassword := AESEncryptKey(sshConnection.PasswordSSH)
		res, err = stmt.Exec(sshConnection.UserSSH, sshConnection.HostSSH, sshConnection.HostNameSSH, sshConnection.PortSSH, sshConnection.CreatorId, encryptedPassword, sshConnection.OsType, sshConnection.IsNetwork, sshConnection.NetworkType, sshConnection.NetworkOS)
	}
	if err != nil {
		return lastId, err
	}
	lastId, err = res.LastInsertId()
	if err != nil {
		return lastId, err
	}

	return lastId, err
}

func GetAllSSHConnection() ([]SshConnectionInfo, error) {
	db := database.ConnectDB()
	defer db.Close()

	query := `SELECT sc_connection_id, sc_username, sc_password, sc_host, sc_hostname, sc_port, creator_id, ssh_key_id, sc_ostype, sc_isnetwork, sc_networkos 
			  FROM ssh_connections`
	selDB, err := db.Query(query)
	if err != nil {
		return nil, err
	}

	var connectionInfo SshConnectionInfo
	var connectionInfos []SshConnectionInfo
	for selDB.Next() {
		var networkOS sql.NullString
		var password sql.NullString
		var keyId sql.NullInt32
		err = selDB.Scan(&connectionInfo.SSHConnectionId, &connectionInfo.UserSSH, &password, &connectionInfo.HostSSH, &connectionInfo.HostNameSSH, &connectionInfo.PortSSH, &connectionInfo.CreatorId, &keyId, &connectionInfo.OsType, &connectionInfo.IsNetwork, &networkOS)
		if err != nil {
			return nil, err
		}
		connectionInfo.NetworkOS = networkOS.String

		// Decrypted Password if exist
		var decryptedPassword string
		if password.String != "" {
			decryptedPassword, err = AESDecryptKey(password.String)
			if err != nil {
				return nil, err
			}

		}
		connectionInfo.PasswordSSH = decryptedPassword
		connectionInfo.SSHKeyId = int(keyId.Int32)
		connectionInfos = append(connectionInfos, connectionInfo)
	}
	return connectionInfos, err
}

func GetAllSSHConnectionNoGroup() ([]SshConnectionInfo, error) {
	db := database.ConnectDB()
	defer db.Close()

	query := `SELECT sc_connection_id, sc_username, sc_password, sc_host, sc_hostname, sc_port, creator_id, ssh_key_id, sc_ostype, sc_isnetwork, sc_networkos 
			  FROM ssh_connections WHERE group_id is null`
	selDB, err := db.Query(query)
	if err != nil {
		return nil, err
	}

	var connectionInfo SshConnectionInfo
	var connectionInfos []SshConnectionInfo
	for selDB.Next() {
		var networkOS sql.NullString
		var password sql.NullString
		var keyId sql.NullInt32
		err = selDB.Scan(&connectionInfo.SSHConnectionId, &connectionInfo.UserSSH, &password, &connectionInfo.HostSSH, &connectionInfo.HostNameSSH, &connectionInfo.PortSSH, &connectionInfo.CreatorId, &keyId, &connectionInfo.OsType, &connectionInfo.IsNetwork, &networkOS)
		if err != nil {
			return nil, err
		}
		connectionInfo.NetworkOS = networkOS.String

		// Decrypted Password if exist
		var decryptedPassword string
		if password.String != "" {
			decryptedPassword, err = AESDecryptKey(password.String)
			if err != nil {
				return nil, err
			}

		}
		connectionInfo.PasswordSSH = decryptedPassword
		connectionInfo.SSHKeyId = int(keyId.Int32)
		connectionInfos = append(connectionInfos, connectionInfo)
	}
	return connectionInfos, err
}

func GetAllSSHConnectionFromGroupId(groupId int) ([]SshConnectionInfo, error) {
	db := database.ConnectDB()
	defer db.Close()

	query := `SELECT sc_connection_id, sc_username, sc_password, sc_host, sc_hostname, sc_port, creator_id, ssh_key_id, sc_ostype, sc_isnetwork, sc_networkos 
			  FROM ssh_connections WHERE group_id = ?`
	selDB, err := db.Query(query, groupId)
	if err != nil {
		return nil, err
	}

	var connectionInfo SshConnectionInfo
	var connectionInfos []SshConnectionInfo
	for selDB.Next() {
		var networkOS sql.NullString
		var password sql.NullString
		var keyId sql.NullInt32
		err = selDB.Scan(&connectionInfo.SSHConnectionId, &connectionInfo.UserSSH, &password, &connectionInfo.HostSSH, &connectionInfo.HostNameSSH, &connectionInfo.PortSSH, &connectionInfo.CreatorId, &keyId, &connectionInfo.OsType, &connectionInfo.IsNetwork, &networkOS)
		if err != nil {
			return nil, err
		}
		connectionInfo.NetworkOS = networkOS.String

		// Decrypted password if exists
		var decryptedPassword string
		if password.String != "" {
			decryptedPassword, err = AESDecryptKey(password.String)
			if err != nil {
				return nil, err
			}
		}
		connectionInfo.PasswordSSH = decryptedPassword
		connectionInfo.SSHKeyId = int(keyId.Int32)
		connectionInfos = append(connectionInfos, connectionInfo)
	}
	return connectionInfos, err
}

func GetAllOSSSHConnection(osType string) ([]SshConnectionInfo, error) {
	db := database.ConnectDB()
	defer db.Close()
	var query string

	if osType == "Linux" {
		query = `SELECT sc_connection_id, sc_username, sc_password, sc_host, sc_hostname, sc_port, creator_id, ssh_key_id, sc_isnetwork, sc_networkos FROM ssh_connections WHERE sc_ostype='Ubuntu' or sc_ostype LIKE '%CentOS%' or sc_ostype LIKE '%Kali%'`
	} else {
		query = `SELECT sc_connection_id, sc_username, sc_password, sc_host, sc_hostname, sc_port, creator_id, ssh_key_id, sc_isnetwork, sc_networkos FROM ssh_connections WHERE sc_ostype LIKE '%Windows%'`
	}
	selDB, err := db.Query(query)
	if err != nil {
		return nil, err
	}

	var connectionInfo SshConnectionInfo
	var connectionInfos []SshConnectionInfo
	for selDB.Next() {
		var networkOS sql.NullString
		var password sql.NullString
		var keyId sql.NullInt32
		err = selDB.Scan(&connectionInfo.SSHConnectionId, &connectionInfo.UserSSH, &password, &connectionInfo.HostSSH, &connectionInfo.HostNameSSH, &connectionInfo.PortSSH, &connectionInfo.CreatorId, &keyId, &connectionInfo.IsNetwork, &networkOS)
		if err != nil {
			return nil, err
		}
		connectionInfo.NetworkOS = networkOS.String

		// Decrypted Password if exist
		var decryptedPassword string
		if password.String != "" {
			decryptedPassword, err = AESDecryptKey(password.String)
			if err != nil {
				return nil, err
			}

		}
		connectionInfo.PasswordSSH = decryptedPassword
		connectionInfo.SSHKeyId = int(keyId.Int32)
		connectionInfos = append(connectionInfos, connectionInfo)
	}
	return connectionInfos, err
}

func GetAllSSHConnectionByNetworkType(networkType string) ([]SshConnectionInfo, error) {
	db := database.ConnectDB()
	defer db.Close()
	var query string

	query = `SELECT sc_connection_id, sc_username, sc_password, sc_host, sc_hostname, sc_port, creator_id, ssh_key_id, sc_isnetwork, sc_networktype, sc_networkos FROM ssh_connections WHERE sc_networktype = ?`
	selDB, err := db.Query(query, networkType)
	if err != nil {
		return nil, err
	}

	var connectionInfo SshConnectionInfo
	var connectionInfos []SshConnectionInfo
	for selDB.Next() {
		var networkOS sql.NullString
		var password sql.NullString
		var keyId sql.NullInt32
		err = selDB.Scan(&connectionInfo.SSHConnectionId, &connectionInfo.UserSSH, &password, &connectionInfo.HostSSH, &connectionInfo.HostNameSSH, &connectionInfo.PortSSH, &connectionInfo.CreatorId, &keyId, &connectionInfo.IsNetwork, &connectionInfo.NetworkType, &networkOS)
		if err != nil {
			return nil, err
		}
		connectionInfo.NetworkOS = networkOS.String

		// Decrypted Password if exist
		var decryptedPassword string
		if password.String != "" {
			decryptedPassword, err = AESDecryptKey(password.String)
			if err != nil {
				return nil, err
			}

		}
		connectionInfo.PasswordSSH = decryptedPassword
		connectionInfo.SSHKeyId = int(keyId.Int32)
		connectionInfos = append(connectionInfos, connectionInfo)
	}
	return connectionInfos, err
}

func GetAllSSHConnectionWithPassword() ([]SshConnectionInfo, error) {
	db := database.ConnectDB()
	defer db.Close()

	query := `SELECT sc_connection_id, sc_username, sc_password, sc_host, sc_hostname, sc_port, creator_id, ssh_key_id, sc_ostype 
			  FROM ssh_connections`
	selDB, err := db.Query(query)
	if err != nil {
		return nil, err
	}

	var connectionInfo SshConnectionInfo
	var connectionInfos []SshConnectionInfo
	for selDB.Next() {
		var password sql.NullString
		var keyId sql.NullInt32
		err = selDB.Scan(&connectionInfo.SSHConnectionId, &connectionInfo.UserSSH, &password, &connectionInfo.HostSSH, &connectionInfo.HostNameSSH, &connectionInfo.PortSSH, &connectionInfo.CreatorId, &keyId, &connectionInfo.OsType)
		if err != nil {
			return nil, err
		}

		// Decrypted Password if exist
		var decryptedPassword string
		if password.String != "" {
			decryptedPassword, err = AESDecryptKey(password.String)
			if err != nil {
				return nil, err
			}

		}
		connectionInfo.PasswordSSH = decryptedPassword
		connectionInfo.SSHKeyId = int(keyId.Int32)

		connectionInfos = append(connectionInfos, connectionInfo)
	}
	return connectionInfos, err
}

func GetSSHConnectionFromHostName(sshHostName string) (*SshConnectionInfo, error) {
	db := database.ConnectDB()
	defer db.Close()

	var sshConnection SshConnectionInfo
	row := db.QueryRow("SELECT sc_connection_id, sc_username, sc_password , sc_host, sc_port, creator_id, ssh_key_id FROM ssh_connections WHERE sc_hostname = ?", sshHostName)
	var password sql.NullString
	var keyId sql.NullInt32
	err := row.Scan(&sshConnection.SSHConnectionId, &sshConnection.UserSSH, &password, &sshConnection.HostSSH, &sshConnection.PortSSH, &sshConnection.CreatorId, &keyId)
	if row == nil {
		return nil, errors.New("ssh connection doesn't exist")
	}
	// Decrypted Password if exist
	var decryptedPassword string
	if password.String != "" {
		decryptedPassword, err = AESDecryptKey(password.String)
		if err != nil {
			return nil, err
		}

	}
	sshConnection.PasswordSSH = decryptedPassword
	sshConnection.SSHKeyId = int(keyId.Int32)

	return &sshConnection, err
}

func GetSSHConnectionFromId(sshConnectionId int) (*SshConnectionInfo, error) {
	db := database.ConnectDB()
	defer db.Close()

	var sshConnection SshConnectionInfo
	row := db.QueryRow("SELECT sc_connection_id, sc_username, sc_password, sc_host, sc_hostname, sc_port, creator_id, ssh_key_id, sc_ostype, sc_networktype, sc_networkos FROM ssh_connections WHERE sc_connection_id = ?", sshConnectionId)
	var password sql.NullString
	var keyId sql.NullInt32
	err := row.Scan(&sshConnection.SSHConnectionId, &sshConnection.UserSSH, &password, &sshConnection.HostSSH, &sshConnection.HostNameSSH, &sshConnection.PortSSH, &sshConnection.CreatorId, &keyId, &sshConnection.OsType, &sshConnection.NetworkType, &sshConnection.NetworkOS)
	if row == nil {
		return nil, errors.New("ssh connection doesn't exist")
	}
	if err != nil {
		return nil, errors.New("fail to retrieve ssh connection info")
	}

	// Decrypted Password if exist
	var decryptedPassword string
	if password.String != "" {
		decryptedPassword, err = AESDecryptKey(password.String)
		if err != nil {
			return nil, err
		}

	}
	sshConnection.PasswordSSH = decryptedPassword
	sshConnection.SSHKeyId = int(keyId.Int32)

	return &sshConnection, err
}

func GetSshHostnameFromId(sshConnectionId int) (string, error) {
	db := database.ConnectDB()
	defer db.Close()

	var hostname string
	row := db.QueryRow("SELECT sc_hostname FROM ssh_connections WHERE sc_connection_id = ?", sshConnectionId)
	err := row.Scan(&hostname)
	if row == nil {
		return hostname, errors.New("ssh connection doesn't exist")
	}

	return hostname, err
}

func GetSSHConnectionFromIP(ip string) (SshConnectionInfo, error) {
	db := database.ConnectDB()
	defer db.Close()

	var sshConnection SshConnectionInfo
	row := db.QueryRow("SELECT sc_connection_id, sc_username, sc_password, sc_host, sc_hostname, sc_port, creator_id, ssh_key_id FROM ssh_connections WHERE sc_host = ?", ip)
	var password sql.NullString
	var keyId sql.NullInt32
	err := row.Scan(&sshConnection.SSHConnectionId, &sshConnection.UserSSH, &password, &sshConnection.HostSSH, &sshConnection.HostNameSSH, &sshConnection.PortSSH, &sshConnection.CreatorId, &keyId)
	if row == nil {
		return sshConnection, errors.New("ssh connection doesn't exist")
	}
	if err != nil {
		return sshConnection, errors.New("fail to retrieve ssh connection info")
	}

	// Decrypted Password if exist
	var decryptedPassword string
	if password.String != "" {
		decryptedPassword, err = AESDecryptKey(password.String)
		if err != nil {
			return sshConnection, err
		}

	}
	sshConnection.PasswordSSH = decryptedPassword
	sshConnection.SSHKeyId = int(keyId.Int32)
	return sshConnection, err
}

// Check Public Key of user exist or not
func (sshConnection *SshConnectionInfo) IsKeyExist() bool {
	if _, err := GetSSHKeyFromId(sshConnection.SSHKeyId); err == nil {
		return true
	} else {
		return false
	}
}

//Delete SSH Connection Function
func DeleteSSHConnection(id int) (bool, error) {
	db := database.ConnectDB()
	defer db.Close()

	stmt, err := db.Prepare("DELETE FROM ssh_connections WHERE sc_connection_id = ?")
	if err != nil {
		return false, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(id)
	if err != nil {
		return false, err
	}
	rows, err := res.RowsAffected()
	if rows == 0 {
		return false, errors.New("no SSH Connections with this ID exists")
	}
	return true, err
}

func GenerateInventory() error {
	var inventory string
	// Get sshConnection No Group
	sshConnectionList, err := GetAllSSHConnectionNoGroup()
	if err != nil {
		return err
	}
	line := GenerateInventoryLine(sshConnectionList)
	inventory += line + "\n"

	groupList, err := GetAllInventoryGroup()
	if err != nil {
		return err
	}
	for _, group := range groupList {
		inventory += "[" + group.GroupName + "]" + "\n"
		sshConnectionList, err = GetAllSSHConnectionFromGroupId(group.GroupId)
		if err != nil {
			return err
		}
		line := GenerateInventoryLine(sshConnectionList)
		inventory += line + "\n"
	}

	err = ioutil.WriteFile("/etc/ansible/hosts", []byte(inventory), 0644)
	return err
}

func GenerateInventoryLine(sshConnectionList []SshConnectionInfo) string {
	var inventory string
	for _, sshConnection := range sshConnectionList {
		var line string

		if sshConnection.IsNetwork {
			line = sshConnection.HostNameSSH + " ansible_host=" + sshConnection.HostSSH + " ansible_port=" + fmt.Sprint(sshConnection.PortSSH) + " ansible_user=" + sshConnection.UserSSH + " ansible_network_os=" + sshConnection.NetworkOS
		} else if strings.Contains(sshConnection.OsType, "Windows") {
			line = sshConnection.HostNameSSH + " ansible_host=" + sshConnection.HostSSH + " ansible_port=" + fmt.Sprint(sshConnection.PortSSH) + " ansible_user=" + sshConnection.UserSSH + " ansible_connection=ssh ansible_shell_type=cmd"
		} else {
			line = sshConnection.HostNameSSH + " ansible_host=" + sshConnection.HostSSH + " ansible_port=" + fmt.Sprint(sshConnection.PortSSH) + " ansible_user=" + sshConnection.UserSSH
		}

		// Append Password if used
		if sshConnection.PasswordSSH != "" {
			line += " ansible_password=" + sshConnection.PasswordSSH
		}
		line += "\n"
		inventory += line
	}
	return inventory
}

//Run command through SSH using SSH keys or Password
func (sshConnection *SshConnectionInfo) RunCommandFromSSHConnectionUseKeys(command string) (string, error) {
	var (
		result string
		err    error
	)

	if sshConnection.PasswordSSH == "" {
		result, err = sshConnection.ExecCommandWithSSHKey(command) // Use Key-Based Authentication
	} else {
		result, err = sshConnection.ExecCommandWithPassword(command) // Use Password Authentication
	}
	return result, err
}

func (sshConnection *SshConnectionInfo) connectSSHWithSSHKeys() (*ssh.Client, error) {
	//If private key is incorrect or wrong format, return error immediately
	var auth []ssh.AuthMethod
	authMethod, err := ProcessPrivateKey(sshConnection.SSHKeyId)
	if err != nil {
		return nil, err
	}
	//Else continue testing connection using the above key
	auth = append(auth, authMethod)

	sshConfig := &ssh.ClientConfig{
		User:            sshConnection.UserSSH,
		Auth:            auth,
		Timeout:         30 * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	addr := fmt.Sprintf("%s:%d", sshConnection.HostSSH, sshConnection.PortSSH)

	sshClient, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return sshClient, err
	}

	return sshClient, nil
}

func (sshConnection *SshConnectionInfo) ExecCommandWithSSHKey(cmd string) (string, error) {

	var (
		session   *ssh.Session
		sshClient *ssh.Client
		err       error
	)

	//create ssh connect
	sshClient, err = sshConnection.connectSSHWithSSHKeys()
	if err != nil {
		return "Wrong username or password to connect remote server", err
	} else {
		defer sshClient.Close()
		//create a session. It is one session per command
		session, err = sshClient.NewSession()
		if err != nil {
			return "Failed to open new session", err
		}
		defer session.Close()

		var b bytes.Buffer //import "bytes"
		var stderr bytes.Buffer
		session.Stdout = &b
		session.Stderr = &stderr
		err = session.Run(cmd)
		if err != nil {
			err = errors.New(fmt.Sprint(fmt.Sprint(err) + ": " + stderr.String()))
		}
		return b.String(), err
	}
}

func (sshConnection *SshConnectionInfo) connectSSHWithPassword() (*ssh.Client, error) {
	var (
		auth         []ssh.AuthMethod
		addr         string
		clientConfig *ssh.ClientConfig
		sshClient    *ssh.Client
		err          error
	)

	// get auth method

	auth = make([]ssh.AuthMethod, 0)
	auth = append(auth, ssh.Password(sshConnection.PasswordSSH))

	clientConfig = &ssh.ClientConfig{
		User:            sshConnection.UserSSH,
		Auth:            auth,
		Timeout:         30 * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// connect to ssh

	addr = fmt.Sprintf("%s:%d", sshConnection.HostSSH, sshConnection.PortSSH)

	sshClient, err = ssh.Dial("tcp", addr, clientConfig)
	if err != nil {
		return sshClient, err
	}

	return sshClient, nil
}

func (sshConnection *SshConnectionInfo) ExecCommandWithPassword(cmd string) (string, error) {

	var (
		session   *ssh.Session
		sshClient *ssh.Client
		err       error
	)

	//create ssh connect
	sshClient, err = sshConnection.connectSSHWithPassword()
	if err != nil {
		return "Wrong username or password to connect remote server", err
	} else {
		defer sshClient.Close()
		//create a session. It is one session per command
		session, err = sshClient.NewSession()
		if err != nil {
			return "Failed to open new session", err
		}
		defer session.Close()
		var b bytes.Buffer //import "bytes"
		session.Stdout = &b
		err = session.Run(cmd)
		return b.String(), err
	}

}

// Get OS Type of PC
func (sshConnection *SshConnectionInfo) GetOsType() string {

	var osType string
	type OsJson struct {
		Name string `json:"name"`
	}
	if sshConnection.IsNetwork {
		return "Unknown"
	}

	output, err := sshConnection.RunCommandFromSSHConnectionUseKeys(`osqueryi --json "SELECT name FROM os_version"`)
	var osJson []OsJson
	if err != nil {
		osType = "Unknown"
	} else {
		err = json.Unmarshal([]byte(output), &osJson)
		if err != nil {
			osType = "Unknown"
		}
		osType = osJson[0].Name
	}
	return osType
}

// Update Os Type to DB
func (sshconnection *SshConnectionInfo) UpdateOsType() error {
	db := database.ConnectDB()
	defer db.Close()

	query := "UPDATE ssh_connections SET sc_ostype = ? WHERE sc_hostname = ?"
	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(sshconnection.OsType, sshconnection.HostNameSSH)
	if err != nil {
		return err
	}
	return err
}

func (sshConnection *SshConnectionInfo) GetWindowsFirewall(direction string) ([]PortNetshFirewallRule, error) {
	var firewallRules []PortNetshFirewallRule
	firewallRule, err := sshConnection.RunCommandFromSSHConnectionUseKeys(`netsh advfirewall firewall show rule name=all dir="` + direction)
	if err != nil {
		return firewallRules, err
	}
	firewallRules, err = ParsePortNetshFirewallRuleFromPowershell(firewallRule)
	return firewallRules, err
}

// Network:Vyos
// List All
func ListAllVyOS() ([]SshConnectionInfo, error) {
	db := database.ConnectDB()
	defer db.Close()

	query := `SELECT sc_connection_id, sc_username, sc_password, sc_host, sc_hostname, sc_port, creator_id, ssh_key_id, sc_isnetwork, sc_networkos
			  FROM ssh_connections WHERE sc_isnetwork=true AND sc_networkos = "vyos"`
	selDB, err := db.Query(query)
	if err != nil {
		return nil, err
	}

	var connectionInfo SshConnectionInfo
	var connectionInfos []SshConnectionInfo
	for selDB.Next() {
		var password sql.NullString
		var keyId sql.NullInt32
		err = selDB.Scan(&connectionInfo.SSHConnectionId, &connectionInfo.UserSSH, &password, &connectionInfo.HostSSH, &connectionInfo.HostNameSSH, &connectionInfo.PortSSH, &connectionInfo.CreatorId, &keyId, &connectionInfo.IsNetwork, &connectionInfo.NetworkOS)
		if err != nil {
			return nil, err
		}

		// Decrypted Password if exist
		var decryptedPassword string
		if password.String != "" {
			decryptedPassword, err = AESDecryptKey(password.String)
			if err != nil {
				return nil, err
			}

		}
		connectionInfo.PasswordSSH = decryptedPassword
		connectionInfo.SSHKeyId = int(keyId.Int32)
		connectionInfos = append(connectionInfos, connectionInfo)
	}
	return connectionInfos, err
}

// Network:Cisco
// List All
func ListAllCisco() ([]SshConnectionInfo, error) {
	db := database.ConnectDB()
	defer db.Close()

	query := `SELECT sc_connection_id, sc_username, sc_password, sc_host, sc_hostname, sc_port, creator_id, ssh_key_id, sc_isnetwork, sc_networkos
			  FROM ssh_connections WHERE sc_isnetwork=true AND sc_networkos = "ios"`
	selDB, err := db.Query(query)
	if err != nil {
		return nil, err
	}

	var connectionInfo SshConnectionInfo
	var connectionInfos []SshConnectionInfo
	for selDB.Next() {
		var password sql.NullString
		var keyId sql.NullInt32
		err = selDB.Scan(&connectionInfo.SSHConnectionId, &connectionInfo.UserSSH, &password, &connectionInfo.HostSSH, &connectionInfo.HostNameSSH, &connectionInfo.PortSSH, &connectionInfo.CreatorId, &keyId, &connectionInfo.IsNetwork, &connectionInfo.NetworkOS)
		if err != nil {
			return nil, err
		}

		// Decrypted Password if exist
		var decryptedPassword string
		if password.String != "" {
			decryptedPassword, err = AESDecryptKey(password.String)
			if err != nil {
				return nil, err
			}

		}
		connectionInfo.PasswordSSH = decryptedPassword
		connectionInfo.SSHKeyId = int(keyId.Int32)

		connectionInfos = append(connectionInfos, connectionInfo)
	}
	return connectionInfos, err
}

func (sshConnection *SshConnectionInfo) GetInstalledProgram() ([]Programs, error) {
	result, err := sshConnection.RunCommandFromSSHConnectionUseKeys(`osqueryi --json "SELECT * FROM programs"`)
	if err != nil {
		return nil, err
	}
	var installedPrograms []Programs

	err = json.Unmarshal([]byte(result), &installedPrograms)
	return installedPrograms, err
}

func (sshConnection SshConnectionInfo) RunAnsiblePlaybookWithjson(filepath string, extraVars string) (string, error) {
	var (
		out, errbuf bytes.Buffer
		err         error
		output      string
	)

	sshKey, err := GetSSHKeyFromId(sshConnection.SSHKeyId)
	if err != nil {
		return "", err
	}
	err = sshKey.WriteKeyToFile("./tmp/private_key_" + strconv.Itoa(sshConnection.SSHKeyId))
	if err != nil {
		return "", err
	}

	var args []string
	args = append(args, "--private-key", "./tmp/private_key_"+strconv.Itoa(sshConnection.SSHKeyId))
	if extraVars != "" {
		args = append(args, "--extra-vars", extraVars, filepath)
	} else {
		args = append(args, filepath)
	}
	defer func() {
		RemoveFile("./tmp/private_key_" + strconv.Itoa(sshConnection.SSHKeyId))
	}()

	cmd := exec.Command("ansible-playbook", args...)
	cmd.Stdout = &out
	cmd.Stderr = &errbuf
	err = cmd.Run()
	stderr := errbuf.String()
	if err != nil {
		// "Exit status 2" means Ansible displays fatal error but our funtion still works correctly
		if err.Error() == "exit status 2" || err.Error() == "exit status 4" {
			err = nil
			log.Println(stderr)
		} else {
			return output, err
		}
	}
	output = out.String()
	return output, err
}

func CountUnknownOS() (int, error) {
	var (
		count int
		err   error
	)
	db := database.ConnectDB()
	defer db.Close()

	// Count sshConnection with os_type is unknown and not a network device
	query := `SELECT count(*) FROM ssh_connections WHERE sc_ostype = "Unknown" and sc_isnetwork = 0`
	selDB, err := db.Query(query)
	if err != nil {
		return count, err
	}

	for selDB.Next() {
		err = selDB.Scan(&count)
		if err != nil {
			return count, err
		}
	}
	return count, err
}

func CountNetworkOS() (int, error) {
	var (
		count int
		err   error
	)
	db := database.ConnectDB()
	defer db.Close()

	// Count sshConnection with os_type is unknown and not a network device
	query := `SELECT count(*) FROM ssh_connections WHERE sc_isnetwork = 1`
	selDB, err := db.Query(query)
	if err != nil {
		return count, err
	}

	for selDB.Next() {
		err = selDB.Scan(&count)
		if err != nil {
			return count, err
		}
	}
	return count, err
}
