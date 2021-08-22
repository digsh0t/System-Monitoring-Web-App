package routes

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

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
	// ipport := fmt.Sprint(r.RemoteAddr)
	// ip := strings.Split(ipport, ":")[0]
	sshConnection, _ := models.GetSSHConnectionFromIP("192.168.163.136")
	err = models.InsertSysInfoToDB(sysInfo, sshConnection.HostSSH, sshConnection.HostNameSSH, sshConnection.SSHConnectionId)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
	}
}
