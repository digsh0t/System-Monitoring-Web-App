package routes

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"io/ioutil"

	"github.com/bitly/go-simplejson"
	"github.com/wintltr/login-api/auth"
	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/utils"
)

// Copy Key to client
func SSHCopyKey(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	if r.Method == "OPTIONS" {
		//CORS
		// return "OKOK"
		json.NewEncoder(w).Encode("OKOK")
		return
	}

	//Authorization
	isAuthorized, err := auth.CheckAuth(r, []string{"admin"})
	if err != nil {
		utils.ERROR(w, http.StatusUnauthorized, errors.New("invalid token").Error())
		return
	}
	if !isAuthorized {
		utils.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized").Error())
		return
	}

	var sshConnectionInfo models.SshConnectionInfo
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("Fail to retrieve ssh connection info with error: %s", err)
	}
	json.Unmarshal(reqBody, &sshConnectionInfo)

	returnJson := simplejson.New()
	isKeyExist := sshConnectionInfo.IsKeyExist()
	if !isKeyExist {
		returnJson.Set("Status", false)
		returnJson.Set("Error", errors.New("your public key does not exist, please generate a pair public and private key").Error())
		utils.JSON(w, http.StatusBadRequest, returnJson)
		return

	} else {

		sshKey, err := models.GetSSHKeyFromId(sshConnectionInfo.SSHKeyId)
		if err != nil {
			utils.ERROR(w, http.StatusBadRequest, err.Error())
			return
		}

		decrypted, err := models.AESDecryptKey(sshKey.PrivateKey)
		if err != nil {
			utils.ERROR(w, http.StatusBadRequest, err.Error())
			return
		}

		data, err := models.GeneratePublicKey([]byte(decrypted))
		if err != nil {
			utils.ERROR(w, http.StatusBadRequest, err.Error())
			return
		}

		publicKey := strings.TrimSuffix(string(data), "\n") + " key" + fmt.Sprint(sshKey.SSHKeyId) + "\n"
		cmd := "echo" + " \"" + publicKey + "\" " + ">> ~/.ssh/authorized_keys"
		models.ExecCommand("mkdir ~/.ssh", sshConnectionInfo.UserSSH, sshConnectionInfo.PasswordSSH, sshConnectionInfo.HostSSH, sshConnectionInfo.PortSSH)
		_, err = models.ExecCommand(cmd, sshConnectionInfo.UserSSH, sshConnectionInfo.PasswordSSH, sshConnectionInfo.HostSSH, sshConnectionInfo.PortSSH)
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
					returnJson.Set("Status", false)
					returnJson.Set("Error", err.Error())
					utils.JSON(w, http.StatusBadRequest, returnJson)
					return
				}

				success, err = sshConnectionInfo.AddSSHConnectionToDB()
				if err != nil {
					returnJson.Set("Status", false)
					returnJson.Set("Error", err.Error())
					utils.JSON(w, http.StatusBadRequest, returnJson)
					return
				}

				err = models.GenerateInventory()
				if err != nil {
					returnJson.Set("Status", false)
					returnJson.Set("Error", errors.New("error while regenerate ansible inventory").Error())
					utils.JSON(w, http.StatusBadRequest, returnJson)
					return
				}

				event := models.Event{
					EventType:   "SSH Connection",
					Description: "SSH Connection to " + sshConnectionInfo.HostNameSSH + " created",
					TimeStampt:  time.Now(),
					CreatorId:   sshConnectionInfo.CreatorId,
				}
				models.CreateEvent(event)

				utils.ReturnInsertJSON(w, success, err)

			}
		} else {
			utils.ERROR(w, http.StatusBadRequest, err.Error())
		}
	}

}
