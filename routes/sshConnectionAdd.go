package routes

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

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
	var eventStatus string
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
			// Get Os Type of PC and update to DB
			sshConnectionInfo.OsType, err = sshConnectionInfo.GetOsType()
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

			/*
				// Update Os Type to DB
				err := sshConnectionInfo.UpdateOsType()
				if err != nil {
					returnJson.Set("Status", false)
					returnJson.Set("Error", errors.New("error while updating os type to database"))
					utils.JSON(w, http.StatusBadRequest, returnJson)
					return
				}
			*/
			// Return Json
			utils.ReturnInsertJSON(w, success, err)
			eventStatus = "successfully"
		}

	}
	// Write Event Web
	description := "Add SSHConnection to " + sshConnectionInfo.HostNameSSH + " " + eventStatus
	_, err = models.WriteWebEvent(r, "SSHConnection", description)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("Fail to write event").Error())
		return
	}

}
