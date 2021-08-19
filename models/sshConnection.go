package models

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/wintltr/login-api/database"
	"github.com/wintltr/login-api/utils"
	"golang.org/x/crypto/ssh"
)

type SshConnectionInfo struct {
	SSHConnectionId int    `json:"sshConnectionId"`
	UserSSH         string `json:"userSSH"`
	PasswordSSH     string `json:"passwordSSH"`
	HostSSH         string `json:"hostSSH"`
	PortSSH         int    `json:"portSSH"`
	CreatorId       int    `json:"creatorId"`
	SSHKeyId        int    `json:"sshKeyId"`
}

//Read private key from private key file
func ProcessPrivateKey(keyId int) (ssh.AuthMethod, error) {
	//buffer, err := ioutil.ReadFile(file)
	//fmt.Println(string(buffer))
	privateKey, _ := GetSSHKeyFromId(keyId)
	decrytedPrivateKey := AESDecryptKey(privateKey.PrivateKey)

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

	addr := fmt.Sprintf("%s:%d", sshConnection.HostSSH, sshConnection.PortSSH)

	_, err = ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return false, err
	} else {
		return true, err
	}
}

func (sshConnection *SshConnectionInfo) AddSSHConnectionToDB() (bool, error) {
	db := database.ConnectDB()
	defer db.Close()

	encryptedPassword := AESEncryptKey(sshConnection.PasswordSSH)

	stmt, err := db.Prepare("INSERT INTO ssh_connections (sc_username, sc_password, sc_host, sc_port, creator_id, ssh_key_id) VALUES (?,?,?,?,?,?)")
	if err != nil {
		return false, err
	}
	defer stmt.Close()

	_, err = stmt.Exec(sshConnection.UserSSH, encryptedPassword, sshConnection.HostSSH, sshConnection.PortSSH, sshConnection.CreatorId, sshConnection.SSHKeyId)
	if err != nil {
		return false, err
	}
	return true, err
}

func (connectionInfo *SshConnectionInfo) GetAllSSHConnection() ([]SshConnectionInfo, error) {
	db := database.ConnectDB()
	defer db.Close()

	query := `SELECT sc_connection_id, sc_username, sc_host, sc_port, creator_id, ssh_key_id 
			  FROM ssh_connections`
	selDB, err := db.Query(query)
	if err != nil {
		return nil, err
	}

	var connectionInfos []SshConnectionInfo
	for selDB.Next() {
		err = selDB.Scan(&connectionInfo.SSHConnectionId, &connectionInfo.UserSSH, &connectionInfo.HostSSH, &connectionInfo.PortSSH, &connectionInfo.CreatorId, &connectionInfo.SSHKeyId)
		if err != nil {
			return nil, err
		}

		connectionInfos = append(connectionInfos, *connectionInfo)
	}
	return connectionInfos, err
}

func GetSSHConnectionFromId(sshConnectionId int) (*SshConnectionInfo, error) {
	db := database.ConnectDB()
	defer db.Close()

	var sshConnection SshConnectionInfo
	var encryptedPassword string
	row := db.QueryRow("SELECT sc_connection_id, sc_username, sc_password, sc_host, sc_port, creator_id, ssh_key_id FROM ssh_connections WHERE sc_connection_id = ?", sshConnectionId)
	err := row.Scan(&sshConnection.SSHConnectionId, &sshConnection.UserSSH, &encryptedPassword, &sshConnection.HostSSH, &sshConnection.PortSSH, &sshConnection.CreatorId, &sshConnection.SSHKeyId)
	if row == nil {
		return nil, errors.New("ssh connection doesn't exist")
	}

	sshConnection.PasswordSSH = AESDecryptKey(encryptedPassword)
	if err != nil {
		return nil, errors.New("fail to retrieve ssh connection info")
	}
	return &sshConnection, err
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

//Get SSH Connection and run command on it
func RunCommandFromSSHConnection(sshConnection SshConnectionInfo, command string) (string, error) {
	result, err := ExecCommand(command, sshConnection.UserSSH, sshConnection.PasswordSSH, sshConnection.HostSSH, sshConnection.PortSSH)
	return result, err
}

func connectSSH(user, password, host string, port int) (*ssh.Client, error) {
	var (
		auth         []ssh.AuthMethod
		addr         string
		clientConfig *ssh.ClientConfig
		sshClient    *ssh.Client
		err          error
	)

	// get auth method

	auth = make([]ssh.AuthMethod, 0)
	auth = append(auth, ssh.Password(password))

	clientConfig = &ssh.ClientConfig{
		User:            user,
		Auth:            auth,
		Timeout:         30 * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// connect to ssh

	addr = fmt.Sprintf("%s:%d", host, port)

	sshClient, err = ssh.Dial("tcp", addr, clientConfig)
	if err != nil {
		return sshClient, err
	}

	return sshClient, nil
}

func ExecCommand(cmd string, userSSH string, passwordSSH string, hostSSH string, portSSH int) (string, error) {

	var (
		session   *ssh.Session
		sshClient *ssh.Client
		err       error
	)

	//create ssh connect
	sshClient, err = connectSSH(userSSH, passwordSSH, hostSSH, portSSH)
	if err != nil {
		return "Wrong username or password to connect remote server", err
	} else {
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
