package routes

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/bitly/go-simplejson"
	"github.com/gorilla/mux"
	"github.com/wintltr/login-api/auth"
	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/utils"
)

func GetWindowsFirewall(w http.ResponseWriter, r *http.Request) {

	//Authorization
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
	direction := vars["direction"]
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to read ssh connection id").Error())
		return
	}
	sshConnection, err := models.GetSSHConnectionFromId(id)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to prepare ssh connection").Error())
		return
	}
	firewallRules, err := sshConnection.GetWindowsFirewall(direction)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to read windows firewall rules from ssh connection").Error())
		return
	}
	utils.JSON(w, http.StatusOK, firewallRules)
}

func AddWindowsFirewall(w http.ResponseWriter, r *http.Request) {

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

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to read windows firewall rules from client windows machine").Error())
		return
	}

	//Firewall type for unmarshalling
	type rule struct {
		SSHConnectionId []int    `json:"ssh_connection_id"`
		RuleName        string   `json:"name"`
		Enabled         string   `json:"enabled"`
		Direction       string   `json:"direction"`
		Profiles        []string `json:"profiles"`
		Grouping        string   `json:"grouping"`
		LocalIP         string   `json:"local_ip"`
		RemoteIP        string   `json:"remote_ip"`
		Protocol        string   `json:"protocol"`
		LocalPort       string   `json:"local_port"`
		RemotePort      string   `json:"remote_port"`
		Action          string   `json:"action"`
	}

	var ru rule
	err = json.Unmarshal(body, &ru)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}

	//Translate ssh connection id list to hostname list
	var hosts []string
	for _, id := range ru.SSHConnectionId {
		sshConnection, err := models.GetSSHConnectionFromId(id)
		if err != nil {
			utils.ERROR(w, http.StatusBadRequest, errors.New("fail to get machine info from provided id").Error())
			return
		}
		hosts = append(hosts, sshConnection.HostNameSSH)
	}

	//Copy unmarshalled rule to rule with right format
	var newRule models.AppliedFirewallRule
	newRule.Host = hosts
	newRule.Action = ru.Action
	newRule.Direction = ru.Direction
	newRule.Enabled = ru.Enabled
	newRule.Grouping = ru.Grouping
	newRule.Profiles = ru.Profiles
	newRule.Protocol = ru.Protocol
	newRule.RuleName = ru.RuleName
	//Optional rule attributes
	if ru.LocalIP == "" {
		newRule.LocalIP = "any"
	} else {
		newRule.LocalIP = ru.LocalIP
	}
	if ru.LocalIP == "" {
		newRule.LocalPort = "any"
	} else {
		newRule.LocalPort = ru.LocalPort
	}
	if ru.LocalIP == "" {
		newRule.RemoteIP = "any"
	} else {
		newRule.RemoteIP = ru.RemoteIP
	}
	if ru.LocalIP == "" {
		newRule.RemotePort = "any"
	} else {
		newRule.RemotePort = ru.RemotePort
	}

	marshalRule, err := json.Marshal(newRule)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to parse new windows firewall rule").Error())
		return
	}

	output, err := models.AddFirewallRule(string(marshalRule))
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}

	// Process Ansible Output
	status, fatal, err := models.ProcessingAnsibleOutput(output)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	returnJson := simplejson.New()
	returnJson.Set("Status", status)
	returnJson.Set("Fatal", fatal)
	utils.JSON(w, http.StatusOK, returnJson)

	// Write Event Web
	hostname, err := models.ConvertListIdToHostnameVer2(ru.SSHConnectionId)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, "fail to get hostname")
		return
	}
	description := "1 windows firewall rule added to host [" + hostname + "] "
	_, err = models.WriteWebEvent(r, "Windows", description)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to write event").Error())
		return
	}

}

func RemoveWindowsFirewallRule(w http.ResponseWriter, r *http.Request) {

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

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to read windows firewall rules from client windows machine").Error())
		return
	}

	//Firewall type for unmarshalling
	type rule struct {
		SSHConnectionId int    `json:"ssh_connection_id"`
		RuleName        string `json:"RuleName"`
		Enabled         string `json:"Enabled"`
		Direction       string `json:"Direction"`
		Profiles        string `json:"Profiles"`
		Grouping        string `json:"Grouping"`
		LocalIP         string `json:"LocalIP"`
		RemoteIP        string `json:"RemoteIP"`
		Protocol        string `json:"Protocol"`
		LocalPort       string `json:"LocalPort"`
		RemotePort      string `json:"RemotePort"`
		Action          string `json:"Action"`
	}

	var ru rule
	err = json.Unmarshal(body, &ru)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}

	//Translate ssh connection id list to hostname list

	sshConnection, err := models.GetSSHConnectionFromId(ru.SSHConnectionId)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to get machine info from provided id").Error())
		return
	}

	//Copy unmarshalled rule to rule with right format
	var newRule models.DeletedFirewallRule
	newRule.Host = sshConnection.HostNameSSH
	newRule.Action = ru.Action
	newRule.Direction = ru.Direction
	newRule.Enabled = ru.Enabled
	newRule.Grouping = ru.Grouping
	newRule.Profiles = ru.Profiles
	newRule.Protocol = ru.Protocol
	newRule.RuleName = ru.RuleName
	newRule.LocalIP = ru.LocalIP
	newRule.LocalPort = ru.LocalPort
	newRule.RemoteIP = ru.RemoteIP
	newRule.RemotePort = ru.RemotePort

	marshalRule, err := json.Marshal(newRule)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to parse delete windows firewall rule").Error())
		return
	}

	output, err := models.DeleteFirewallRule(string(marshalRule))
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to delete firewall rule from client windows machine").Error())
		return
	}

	// Process Ansible Output
	status, fatal, err := models.ProcessingAnsibleOutput(output)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	returnJson := simplejson.New()
	returnJson.Set("Status", status)
	returnJson.Set("Fatal", fatal)
	utils.JSON(w, http.StatusOK, returnJson)

	// Write Event Web
	description := "1 windows firewall rule removed from host [" + sshConnection.HostNameSSH + "] "
	_, err = models.WriteWebEvent(r, "Windows", description)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to write event").Error())
		return
	}
}

func GetWindowsOpenConnection(w http.ResponseWriter, r *http.Request) {

	//Authorization
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
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	sshConnection, err := models.GetSSHConnectionFromId(id)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	connectionList, err := sshConnection.GeteOpenConnection()
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	utils.JSON(w, http.StatusOK, connectionList)
}
