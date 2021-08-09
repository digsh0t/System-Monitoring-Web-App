package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/wintltr/login-api/routes"
)

func main() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/login", routes.Login).Methods("POST")

	router.HandleFunc("/sshconnection", routes.SSHCopyKey).Methods("POST")
	router.HandleFunc("/sshconnection/{id}/test", routes.TestSSHConnection).Methods("GET")
	router.HandleFunc("/sshconnections", routes.GetAllSSHConnection).Methods("GET")
	router.HandleFunc("/sshconnection/update", routes.UpdateSSHConnection).Methods("POST")

	router.HandleFunc("/sshkey", routes.AddSSHKey).Methods("POST")
	router.HandleFunc("/sshkeys", routes.GetAllSSHKey).Methods("GET")
	log.Fatal(http.ListenAndServe(":8081", router))
}
