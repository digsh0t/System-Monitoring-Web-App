package models

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/wintltr/login-api/database"
	"github.com/wintltr/login-api/utils"
	"golang.org/x/crypto/ssh"
)

type SshConnectionInfo struct {
	UserSSH     string `json:"userSSH"`
	PasswordSSH string `json:"passwordSSH"`
	HostSSH     string `json:"hostSSH"`
	PortSSH     int    `json:"portSSH"`
	CreatorId   int    `json:"creatorId"`
	SSHKeyId    int    `json:"sshKeyId"`
}

//Test SSH connection using username and password
func (sshConnection *SshConnectionInfo) TestConnection() (bool, error) {
	sshConfig := &ssh.ClientConfig{
		User: sshConnection.UserSSH,
		Auth: []ssh.AuthMethod{
			ssh.Password(sshConnection.PasswordSSH),
		},
		Timeout:         30 * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	addr := fmt.Sprintf("%s:%d", sshConnection.HostSSH, sshConnection.PortSSH)

	_, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return false, err
	} else {
		return true, err
	}
}

//Read public key from private key file
func ReadPublicKeyFile(file string) (ssh.AuthMethod, error) {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

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
	authMethod, err := ReadPublicKeyFile("/home/wintltr/.ssh/id_rsa")
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

	stmt, err := db.Prepare("INSERT INTO ssh_connections (sc_username, sc_host, sc_port, creator_id, ssh_key_id) VALUES (?,?,?,?,?)")
	if err != nil {
		return false, err
	}

	_, err = stmt.Exec(sshConnection.UserSSH, sshConnection.HostSSH, sshConnection.PortSSH, sshConnection.CreatorId, sshConnection.SSHKeyId)
	if err != nil {
		return false, err
	}
	return true, err
}
