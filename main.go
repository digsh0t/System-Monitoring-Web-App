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

	router.HandleFunc("/sshconnection/{id}", routes.SSHConnectionDeleteRoute).Methods("DELETE")

	router.HandleFunc("/sshkey", routes.AddSSHKey).Methods("POST")
	router.HandleFunc("/sshkey/{id}", routes.SSHKeyDeleteRoute).Methods("DELETE")
	router.HandleFunc("/sshkeys", routes.GetAllSSHKey).Methods("GET")

	//Network Function
	router.HandleFunc("/network/defaultip", routes.GetAllDefaultIP).Methods("GET")

	// Load file yaml
	router.HandleFunc("/yamls/{name}", routes.LoadFile).Methods("GET")

	log.Fatal(http.ListenAndServe(":8081", router))
}
