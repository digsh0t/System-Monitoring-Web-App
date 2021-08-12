package routes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"io/ioutil"

	"github.com/bitly/go-simplejson"
	"github.com/wintltr/login-api/auth"
	"github.com/wintltr/login-api/models"
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

func execCommand(cmd string, userSSH string, passwordSSH string, hostSSH string, portSSH int) (string, error) {

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

// Copy Key to client
func SSHCopyKey(w http.ResponseWriter, r *http.Request) {
	//Authorization
	isAuthorized, err := auth.CheckAuth(r, []string{"admin"})
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	if !isAuthorized {
		utils.ERROR(w, http.StatusUnauthorized, err.Error())
		return
	}

	var sshConnectionInfo models.SshConnectionInfo
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("Fail to retrieve ssh connection info with error: %s", err)
	}
	json.Unmarshal(reqBody, &sshConnectionInfo)

	isKeyExist := sshConnectionInfo.IsKeyExist()
	if !isKeyExist {
		utils.ERROR(w, http.StatusNotFound, "Your public key does not exist, please generate a pair public and private key!")

	} else {
		returnJson := simplejson.New()
		sshKey, err := models.GetSSHKeyFromId(sshConnectionInfo.SSHKeyId)
		if err != nil {
			returnJson.Set("Status", 400)
			returnJson.Set("Error", err.Error())
			utils.JSON(w, http.StatusBadRequest, returnJson)
			return
		}

		decrypted := models.AESDecryptKey(sshKey.PrivateKey)
		data, err := models.GeneratePublicKey([]byte(decrypted))
		if err != nil {
			returnJson.Set("Status", 400)
			returnJson.Set("Error", err.Error())
			utils.JSON(w, http.StatusBadRequest, returnJson)
			return
		}

		cmd := "echo" + " \"" + string(data) + "\" " + ">> ~/.ssh/authorized_keys"
		message, err := execCommand(cmd, sshConnectionInfo.UserSSH, sshConnectionInfo.PasswordSSH, sshConnectionInfo.HostSSH, sshConnectionInfo.PortSSH)
		if err == nil {

			//Test the SSH connection using public key if works
			success, err := sshConnectionInfo.TestConnectionPublicKey()
			if err != nil {
				returnJson.Set("Status", success)
				returnJson.Set("Error", err.Error())
				utils.JSON(w, http.StatusBadRequest, returnJson)
			} else {

				sshConnectionInfo.CreatorId, err = auth.ExtractUserId(r)
				if err != nil {
					returnJson.Set("Status", 400)
					returnJson.Set("Error", err.Error())
					utils.JSON(w, http.StatusBadRequest, returnJson)
					return
				}

				success, err = sshConnectionInfo.AddSSHConnectionToDB()
				utils.ReturnInsertJSON(w, success, err)
			}
		} else {
			utils.ERROR(w, http.StatusBadRequest, message)
		}
	}

}
