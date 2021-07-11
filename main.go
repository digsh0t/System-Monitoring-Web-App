package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/wintltr/login-api/routes"
	sshConnect "github.com/wintltr/login-api/ssh_connect"
)

func main() {
	sshConnect.ConnectWithPassword("dscmember", "dsc@2021", "192.168.163.136", "22")
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/login", routes.Login).Methods("POST")
	log.Fatal(http.ListenAndServe(":8081", router))
}
