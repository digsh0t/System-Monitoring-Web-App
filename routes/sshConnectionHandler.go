package routes

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/bitly/go-simplejson"
	"github.com/gorilla/mux"
	"github.com/wintltr/login-api/auth"
	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/utils"
)

func TestSSHConnection(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	if r.Method == "OPTIONS" {
		//CORS
		// return "OKOK"
		json.NewEncoder(w).Encode("OKOK")
		return
	}

	isAuthorized, err := auth.CheckAuth(r, []string{"admin", "user"})
	if err != nil {
		utils.ERROR(w, http.StatusUnauthorized, errors.New("invalid token").Error())
		return
	}
	if !isAuthorized {
		utils.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized").Error())
		return
	}

	vars := mux.Vars(r)
	sshConnectionId, _ := strconv.Atoi(vars["id"])
	sshConnection, err := models.GetSSHConnectionFromId(sshConnectionId)

	status := false
	returnJson := simplejson.New()

	if err != nil {
		returnJson.Set("Status", status)
		returnJson.Set("Error", err.Error())
		utils.JSON(w, http.StatusBadRequest, returnJson)
		return
	}

	status, err = sshConnection.TestConnectionPublicKey()
	var eventStatus string
	if err != nil {
		returnJson.Set("Status", status)
		returnJson.Set("Error", err.Error())
		utils.JSON(w, http.StatusBadRequest, returnJson)
		eventStatus = "failed"
	} else {
		returnJson.Set("Status", status)
		returnJson.Set("Error", "")
		utils.JSON(w, http.StatusOK, returnJson)
		eventStatus = "successfully"
	}

	// Write Event Web
	description := "Test SSHconnection " + sshConnection.HostNameSSH + " " + eventStatus
	_, err = models.WriteWebEvent(r, "SSHConnection", description)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("Fail to write event").Error())
		return
	}
}

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
			sshConnectionInfo.OsType = sshConnectionInfo.GetOsType()

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

func SSHConnectionDeleteRoute(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	if r.Method == "OPTIONS" {
		//CORS
		// return "OKOK"
		json.NewEncoder(w).Encode("OKOK")
		return
	}

	isAuthorized, err := auth.CheckAuth(r, []string{"admin"})
	if err != nil {
		utils.ERROR(w, http.StatusUnauthorized, errors.New("invalid token").Error())
		return
	}
	if !isAuthorized {
		utils.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized").Error())
		return
	}

	returnJson := simplejson.New()
	vars := mux.Vars(r)
	sshConnectionId, err := strconv.Atoi(vars["id"])
	if err != nil {
		returnJson.Set("Status", false)
		returnJson.Set("Error", errors.New("invalid SSH Connection id").Error())
		utils.JSON(w, http.StatusBadRequest, returnJson)
		return
	}

	sshConnectionInfo, _ := models.GetSSHConnectionFromId(sshConnectionId)
	if sshConnectionInfo == nil {
		returnJson.Set("Status", false)
		returnJson.Set("Error", errors.New("SSH Connection with id "+strconv.Itoa(sshConnectionId)+" doesn't exist").Error())
		utils.JSON(w, http.StatusBadRequest, returnJson)
		return
	}

	_, err = models.DeleteSSHConnection(sshConnectionId)
	if err != nil {
		returnJson.Set("Status", false)
		returnJson.Set("Error", errors.New("error while deleting SSH Connection").Error())
		utils.JSON(w, http.StatusBadRequest, returnJson)
		return
	}
	err = models.GenerateInventory()
	var eventStatus string
	if err != nil {
		returnJson.Set("Status", false)
		returnJson.Set("Error", errors.New("error while regenerate ansible inventory").Error())
		utils.JSON(w, http.StatusBadRequest, returnJson)
		eventStatus = "failed"
	} else {
		returnJson.Set("Status", true)
		returnJson.Set("Error", nil)
		utils.JSON(w, http.StatusOK, returnJson)
		eventStatus = "successfully"
	}

	// Write Event Web
	description := "Delete SSHconnection from " + sshConnectionInfo.HostNameSSH + " " + eventStatus
	_, err = models.WriteWebEvent(r, "SSHConnection", description)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("Fail to write event").Error())
		return
	}
}

// Get SSh connection from DB
func GetAllSSHConnection(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	if r.Method == "OPTIONS" {
		//CORS
		// return "OKOK"
		json.NewEncoder(w).Encode("OKOK")
		return
	}

	isAuthorized, err := auth.CheckAuth(r, []string{"admin", "user"})
	if err != nil {
		utils.ERROR(w, http.StatusUnauthorized, errors.New("invalid token").Error())
		return
	}
	if !isAuthorized {
		utils.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized").Error())
		return
	}

	vars := mux.Vars(r)
	osType := vars["ostype"]
	if osType == "" {
		sshConnectionList, err := models.GetAllSSHConnection()
		if err != nil {
			utils.JSON(w, http.StatusBadRequest, err.Error())
			return
		}
		utils.JSON(w, http.StatusOK, sshConnectionList)
		return
	}

	sshConnectionList, err := models.GetAllOSSSHConnection(osType)
	if err != nil {
		utils.JSON(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, sshConnectionList)

}

// Get SSh connection from DB
func GetAllSSHConnectionNoGroup(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	if r.Method == "OPTIONS" {
		//CORS
		// return "OKOK"
		json.NewEncoder(w).Encode("OKOK")
		return
	}

	isAuthorized, err := auth.CheckAuth(r, []string{"admin", "user"})
	if err != nil {
		utils.ERROR(w, http.StatusUnauthorized, errors.New("invalid token").Error())
		return
	}
	if !isAuthorized {
		utils.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized").Error())
		return
	}

	sshConnectionList, err := models.GetAllSSHConnection()
	if err != nil {
		utils.JSON(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, sshConnectionList)

}

func AddNewSSHConnection(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	log.Println(strings.Split(r.RemoteAddr, ":")[0])

	type receivedInfo struct {
		UserSSH   string `json:"user_ssh"`
		Hostname  string `json:"hostname"`
		IP        string `json:"ip"`
		Port      int    `json:"port"`
		PublicKey string `json:"public_key"`
		OSType    string `json:"os_type"`
		IsNetwork bool   `json:"is_network"`
		NetworkOS string `json:"network_os"`
	}
	var info receivedInfo
	var sshConnection models.SshConnectionInfo

	sshConnection.CreatorId, err = auth.ExtractUserId(r)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}

	err = json.Unmarshal(body, &info)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	//Get SSH Key id
	sshConnection.SSHKeyId, err = models.GetKeyIdFromPublicKey(info.PublicKey)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	//Add received info to ssh connection
	sshConnection.HostNameSSH = info.Hostname
	//sshConnection.HostNameSSH = "vmware-ubuntu"
	sshConnection.HostSSH = strings.Split(r.RemoteAddr, ":")[0]
	//sshConnection.HostSSH = "192.168.163.139"
	sshConnection.PortSSH = info.Port
	sshConnection.OsType = info.OSType
	sshConnection.UserSSH = info.UserSSH
	sshConnection.IsNetwork = info.IsNetwork
	success, err := sshConnection.TestConnectionPublicKey()
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	if success {
		_, err := sshConnection.AddSSHConnectionToDB()
		if err != nil {
			utils.ERROR(w, http.StatusBadRequest, err.Error())
			return
		}
		err = models.GenerateInventory()
		if err != nil {
			utils.ERROR(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	utils.JSON(w, http.StatusOK, body)
}
