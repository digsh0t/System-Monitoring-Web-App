package routes

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/wintltr/login-api/auth"
	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/utils"
)

func AddNewSSHConnection(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}

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
	sshConnection.HostSSH = info.IP
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
