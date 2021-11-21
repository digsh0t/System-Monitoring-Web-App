package routes

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/wintltr/login-api/auth"
	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/utils"
	"github.com/wintltr/login-api/webconsole"
	"golang.org/x/crypto/ssh"
)

var (
	//Allow cross-domain
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func WebConsoleWSHanlder(w http.ResponseWriter, r *http.Request) {

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
	var (
		conn    *websocket.Conn
		client  *ssh.Client
		sshConn *webconsole.SSHConnect
	)
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])
	// password = vars["password"]
	// host = vars["ip"]
	// port, _ = strconv.Atoi(vars["port"])
	if conn, err = upgrader.Upgrade(w, r, nil); err != nil {
		return
	}
	defer conn.Close()

	sshConnection, err := models.GetSSHConnectionFromId(id)
	if err != nil {
		webconsole.WsSendText(conn, []byte(errors.New("SSH Connection id is not available").Error()))
		return
	}
	//Create ssh client
	if sshConnection.PasswordSSH == "" {
		client, err = sshConnection.ConnectSSHWithSSHKeys()
	} else {
		client, err = sshConnection.ConnectSSHWithPassword()
	}
	if err != nil {
		webconsole.WsSendText(conn, []byte(err.Error()))
		return
	}
	defer client.Close()

	//connect to ssh
	if sshConn, err = webconsole.NewSSHConnect(client); err != nil {
		webconsole.WsSendText(conn, []byte(err.Error()))
		return
	}

	quit := make(chan int)
	go sshConn.Output(conn, quit)
	go sshConn.Recv(conn, quit)
	<-quit
}

func WebConsoleTemplate(w http.ResponseWriter, r *http.Request) {

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
	type sshID struct {
		ID string
	}

	var sI sshID
	vars := mux.Vars(r)
	sI.ID = vars["id"]
	temp, e := template.ParseFiles("./template/index.html")
	if e != nil {
		fmt.Println(e)
	}
	temp.Execute(w, sI)
	return
}
