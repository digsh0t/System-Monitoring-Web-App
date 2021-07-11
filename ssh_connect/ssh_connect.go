package sshConnect

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

func ConnectSSH(user, password, host string, port int) (*ssh.Client, error) {
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
		fmt.Printf("Error while connect to SSH client: %s", err)
	}

	return sshClient, nil
}

func ConnectSFTP(sshClient *ssh.Client) (*sftp.Client, error) {

	var (
		sftpClient *sftp.Client
		err        error
	)

	//create sftp client
	sftpClient, err = sftp.NewClient(sshClient)
	if err != nil {
		fmt.Printf("Error while connect to SFTP client: %s", err)
	}

	return sftpClient, nil
}

func SendFile(userSSH string, passwordSSH string, hostSSH string, portSSH int) {
	var (
		err        error
		sftpClient *sftp.Client
		sshClient  *ssh.Client
	)

	sshClient, err = ConnectSSH(userSSH, passwordSSH, hostSSH, portSSH)
	if err != nil {
		fmt.Printf("Error while connect SSH to send file: %s", err)
	}

	sftpClient, err = ConnectSFTP(sshClient)
	if err != nil {
		fmt.Printf("Error while connect to SFTP client: %s", err)
	}
	defer sftpClient.Close()

	//Local file path and folder on remote machine for testing

	var localFilePath = "/tmp/pl.yml"
	var remoteDir = "/tmp/ansibleFile"

	srcFile, err := os.Open(localFilePath)
	if err != nil {
		fmt.Printf("Error while open file: %s", err)
	}
	defer srcFile.Close()

	var remoteFileName = path.Base(localFilePath)
	dstFile, err := sftpClient.Create(path.Join(remoteDir, remoteFileName))
	if err != nil {
		fmt.Printf("Error while write file to remote server: %s", err)
	}
	defer dstFile.Close()

	buf := make([]byte, 1024)
	for {
		n, _ := srcFile.Read(buf)
		if n == 0 {
			break
		}
		dstFile.Write(buf[0:n])
	}

	fmt.Println("Copy file to remote server finished!")

}
