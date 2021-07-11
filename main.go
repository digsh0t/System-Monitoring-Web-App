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
	router.HandleFunc("/ssh-copy-id", routes.SSHCopyKey).Methods("GET")
	log.Fatal(http.ListenAndServe(":8081", router))
}
