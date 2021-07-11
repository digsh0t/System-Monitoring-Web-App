package routes

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"time"

	"io/ioutil"

	"github.com/bitly/go-simplejson"
	"github.com/wintltr/login-api/utils"
	"golang.org/x/crypto/ssh"
)

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

func execCommand(cmd string) (string, error) {

	var (
		session   *ssh.Session
		sshClient *ssh.Client
		err       error
	)

	// Input from user
	var (
		user_SSH     string = "root"
		password_SSH string = "P@ssword"
		host_SSH     string = "192.168.45.100"
		port_SSH     int    = 22
	)

	//create ssh connect
	sshClient, err = connectSSH(user_SSH, password_SSH, host_SSH, port_SSH)
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

// Check Public Key of user exist or not
func isKeyExist() bool {
	user := utils.GetCurrentUser()
	if _, err := os.Stat(user.HomeDir + "/.ssh/id_rsa.pub"); err == nil {
		return true
	} else {
		return false
	}
}

// Copy Key to remote server
func SSHCopyKey(w http.ResponseWriter, r *http.Request) {
	isKeyExist := isKeyExist()
	user := utils.GetCurrentUser()
	if isKeyExist == false {
		utils.ERROR(w, http.StatusNotFound, "Your public key does not exist")

	} else {
		data, _ := ioutil.ReadFile(user.HomeDir + "/.ssh/id_rsa.pub")
		cmd := "echo" + " \"" + string(data) + "\" " + ">> ~/.ssh/authorized_keys"
		message, err := execCommand(cmd)
		if err == nil {
			returnJson := simplejson.New()
			returnJson.Set("Message", "Transfer key successfully!")
			returnJson.Set("Status", true)
			returnJson.Set("Error", "")
			utils.JSON(w, http.StatusOK, returnJson)
		} else {
			utils.ERROR(w, http.StatusBadRequest, message)
		}

	}

}
