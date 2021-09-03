package routes

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/wintltr/login-api/auth"
	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/utils"
)

func DeleteUfwRule(w http.ResponseWriter, r *http.Request) {

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

	var body struct {
		SshConnectionId int    `json:"sshconnectionid"`
		FromIP          string `json:"fromip"`
		FromPort        int    `json:"fromport"`
		ToIP            string `json:"toip"`
		ToPort          int    `json:"toport"`
		Protocol        string `json:"protocol"`
	}
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to read request body").Error())
		return
	}
	err = json.Unmarshal(reqBody, &body)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("invalid ufw rule data").Error())
		return
	}

	sshConnection, err := models.GetSSHConnectionFromId(body.SshConnectionId)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("ssh connection does not exist").Error())
		return
	}

	vars := map[string]interface{}{
		"host":      sshConnection.HostNameSSH,
		"from_ip":   body.FromIP,
		"from_port": body.FromPort,
		"to_ip":     body.ToIP,
		"to_port":   body.ToPort,
		"protocol":  body.Protocol,
	}
	err = models.RunAnsiblePlaybookWithVars(vars, "yamls/delete_ufw_rule.yml")
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("ansible_playbook for deleting ufw rule is missing").Error())
		return
	}
	utils.JSON(w, http.StatusOK, body)
}
