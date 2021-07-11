package models

import (
	"fmt"
	"time"

	"golang.org/x/crypto/ssh"
)

type SshConnectionInfo struct {
	UserSSH     string `json:"userSSH"`
	PasswordSSH string `json:"passwordSSH"`
	HostSSH     string `json:"hostSSH"`
	PortSSH     int    `json:"portSSH"`
}

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
