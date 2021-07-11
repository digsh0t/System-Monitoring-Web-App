package routes

import (
	"bytes"
	"fmt"
	"log"
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
		log.Fatal("Error at sshClient!")
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

	//create a session. It is one session per command
	session, err = sshClient.NewSession()
	if err != nil {
		log.Fatal("Error at open session for ssh!")
	}
	defer session.Close()
	var b bytes.Buffer //import "bytes"
	session.Stdout = &b
	err = session.Run(cmd)
	return b.String(), err
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
		log.Fatal("Your key does not exist")
	}
	data, err := ioutil.ReadFile(user.HomeDir + "/.ssh/id_rsa.pub")
	if err != nil {
		utils.ERROR(w, http.StatusNotFound, "Failed to load public key of your folder!")
	}
	cmd := "echo" + " \"" + string(data) + "\" " + ">> ~/.ssh/authorized_keys"
	execCommand(cmd)
	returnJson := simplejson.New()
	returnJson.Set("Message", "Transfer key successfully!")
	returnJson.Set("Status", true)
	returnJson.Set("Error", "")
	utils.JSON(w, http.StatusOK, returnJson)

}
