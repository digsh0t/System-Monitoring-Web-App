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
	"github.com/wintltr/login-api/notifications"
	"github.com/wintltr/login-api/utils"
)

type localUser struct {
	SSHConnectionId          []int    `json:"ssh_connection_id"`
	AccountDisabled          string   `json:"account_disabled"`
	Description              string   `json:"description"`
	Fullname                 string   `json:"fullname"`
	Group                    []string `json:"group"`
	HomeDirectory            string   `json:"home_directory"`
	LoginScript              string   `json:"login_script"`
	Username                 string   `json:"username"`
	Password                 string   `json:"password"`
	PasswordExpired          string   `json:"password_expired"`
	PasswordNeverExpires     string   `json:"password_never_expires"`
	Profile                  string   `json:"profile"`
	UserCannotChangePassword string   `json:"user_cannot_change_password"`
}

func GetWindowsLocalUser(w http.ResponseWriter, r *http.Request) {

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
	userList, err := sshConnection.GetLocalUsers()
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	utils.JSON(w, http.StatusOK, userList)
}

func AddNewWindowsLocalUser(w http.ResponseWriter, r *http.Request) {

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
		utils.ERROR(w, http.StatusOK, err.Error())
		return
	}
	var lu localUser
	err = json.Unmarshal(body, &lu)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}

	//Translate ssh connection id list to hostname list
	var hosts []string
	for _, id := range lu.SSHConnectionId {
		sshConnection, err := models.GetSSHConnectionFromId(id)
		if err != nil {
			utils.ERROR(w, http.StatusBadRequest, err.Error())
			return
		}
		hosts = append(hosts, sshConnection.HostNameSSH)
	}

	var user models.NewLocalUser
	user.Host = hosts
	user.AccountDisabled = lu.AccountDisabled
	user.Description = lu.Description
	user.Fullname = lu.Fullname
	user.Group = lu.Group
	user.HomeDirectory = lu.HomeDirectory
	user.LoginScript = lu.LoginScript
	user.Password = lu.Password
	if lu.PasswordExpired == "" {
		user.PasswordExpired = "yes"
	} else {
		user.PasswordExpired = lu.PasswordExpired
	}
	if lu.PasswordNeverExpires == "" {
		user.PasswordNeverExpires = "no"
	} else {
		user.PasswordNeverExpires = lu.PasswordNeverExpires
	}
	user.Profile = lu.Profile
	if lu.UserCannotChangePassword == "" {
		user.UserCannotChangePassword = "no"
	} else {
		user.UserCannotChangePassword = lu.UserCannotChangePassword
	}
	user.Username = lu.Username
	marshalledUser, _ := json.Marshal(user)
	output, err := models.AddNewWindowsUser(string(marshalledUser))
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	status, fatal, err := models.ProcessingAnsibleOutput(output)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	returnJson := simplejson.New()
	returnJson.Set("Status", status)
	returnJson.Set("Fatal", fatal)
	utils.JSON(w, http.StatusOK, returnJson)

}

func DeleteWindowsUser(w http.ResponseWriter, r *http.Request) {

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
		utils.ERROR(w, http.StatusOK, err.Error())
		return
	}

	//User type for unmarshalling
	type unmarshalledUser struct {
		SSHConnectionId int    `json:"ssh_connection_id"`
		Name            string `json:"username"`
	}

	type deletedUser struct {
		Host string `json:"host"`
		Name string `json:"username"`
	}

	var uu unmarshalledUser
	err = json.Unmarshal(body, &uu)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}

	//Translate ssh connection id list to hostname list
	sshConnection, err := models.GetSSHConnectionFromId(uu.SSHConnectionId)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}

	var dU deletedUser
	dU.Host = sshConnection.HostNameSSH
	dU.Name = uu.Name
	marshalled, err := json.Marshal(dU)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	output, err := models.DeleteWindowsUser(string(marshalled))
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
}

func GetWindowsGroupListOfUser(w http.ResponseWriter, r *http.Request) {

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

	type groupList struct {
		List []string `json:"groupname"`
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	username := vars["username"]
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	sshConnection, err := models.GetSSHConnectionFromId(id)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	groupNameList, err := sshConnection.GetWindowsGroupUserBelongTo(username)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, groupList{groupNameList})
}

func ReplaceWindowsGroupOfUser(w http.ResponseWriter, r *http.Request) {

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
		utils.ERROR(w, http.StatusOK, err.Error())
		return
	}

	//User type for unmarshalling
	type replacedGroupList struct {
		SSHConnectionId int      `json:"ssh_connection_id"`
		Name            string   `json:"username"`
		Group           []string `json:"groupname"`
	}

	var rGL replacedGroupList
	err = json.Unmarshal(body, &rGL)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}

	//Translate ssh connection id list to hostname list
	sshConnection, err := models.GetSSHConnectionFromId(rGL.SSHConnectionId)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	output, err := sshConnection.ReplaceWindowsGroupForUser(rGL.Name, rGL.Group)
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
}

func KillWindowsLogonSession(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	sessionId, err := strconv.Atoi(vars["session_id"])
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	sshConnection, err := models.GetSSHConnectionFromId(id)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	err = sshConnection.KillWindowsLoginSession(sessionId)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
}

func GetWindowsLogonAppExecutionHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	username := vars["username"]
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	sshConnection, err := models.GetSSHConnectionFromId(id)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	appHistory, err := sshConnection.GetWindowsLoginAppExecutionHistory(username)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	utils.JSON(w, http.StatusOK, appHistory)
}

// func GetWindowsUserEnableStatus(w http.ResponseWriter, r *http.Request) {

// 	vars := mux.Vars(r)
// 	id, err := strconv.Atoi(vars["id"])
// 	if err != nil {
// 		utils.ERROR(w, http.StatusBadRequest, err.Error())
// 		return
// 	}
// 	username := vars["username"]
// 	sshConnection, err := models.GetSSHConnectionFromId(id)
// 	if err != nil {
// 		utils.ERROR(w, http.StatusBadRequest, err.Error())
// 		return
// 	}
// 	isEnabled, err := sshConnection.CheckIfWindowsUserEnabled(username)
// 	if err != nil {
// 		utils.ERROR(w, http.StatusBadRequest, err.Error())
// 		return
// 	}
// 	returnJson := simplejson.New()
// 	returnJson.Set("username", username)
// 	returnJson.Set("is_enabled", isEnabled)

// 	utils.JSON(w, http.StatusOK, returnJson)
// }

func ChangeWindowsEnabledStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	username := vars["username"]
	isEnabled, err := strconv.ParseBool(vars["is_enabled"])
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	sshConnection, err := models.GetSSHConnectionFromId(id)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	err = sshConnection.ChangeWindowsUserEnableStatus(username, isEnabled)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
}

func ChangeWindowsLocalUserPassword(w http.ResponseWriter, r *http.Request) {

	type newPassword struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	userid, err := auth.ExtractUserId(r)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}

	var nP newPassword
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	err = json.Unmarshal(body, &nP)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	sshConnection, err := models.GetSSHConnectionFromId(id)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
	err = sshConnection.ChangeWindowsLocalUserPassword(nP.Username, nP.Password)
	if err != nil {
		// utils.ERROR(w, http.StatusBadRequest, err.Error())
		// return
		notifications.SendToNotificationChannel(err.Error(), "notification-"+strconv.Itoa(userid)+"channel", "change-windows-user-password")
	}
}
