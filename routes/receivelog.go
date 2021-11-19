package routes

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/utils"
)

func Receivelog(w http.ResponseWriter, r *http.Request) {
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("Fail to retrieve ssh connection info with error: %s", err)
	}
	var sysInfo models.SysInfo
	json.Unmarshal(reqBody, &sysInfo)
	sysInfo.Timestamp = time.Now().Format("01-02-2006 15:04:05")
	ipport := fmt.Sprint(r.RemoteAddr)
	ip := strings.Split(ipport, ":")[0]
	sshConnection, _ := models.GetSSHConnectionFromIP(ip)

	/*
		err = models.InsertUfwToDB(sshConnection.HostNameSSH, sshConnection.SSHConnectionId, sysInfo.UfwStatus, sysInfo.UfwRules)
		if err != nil {
			utils.ERROR(w, http.StatusBadRequest, err.Error())
			return
		}
	*/
	err = models.InsertSysInfoToDB(sysInfo, sshConnection.HostSSH, sshConnection.HostNameSSH, sshConnection.SSHConnectionId)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}

}
