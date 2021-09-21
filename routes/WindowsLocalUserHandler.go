package routes

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/wintltr/login-api/models"
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
	err = models.AddNewWindowsUser(string(marshalledUser))
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
}

func DeleteWindowsUser(w http.ResponseWriter, r *http.Request) {
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
	err = models.DeleteWindowsUser(string(marshalled))
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, err.Error())
		return
	}
}
