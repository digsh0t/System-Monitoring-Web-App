package main

import (
	"log"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/routes"
)

func main() {
	//go goroutines.CheckClientOnlineStatusGour()
	// firewallRule, err := sshConnection.RunCommandFromSSHConnectionUseKeys(`osqueryi --json "SELECT * FROM iptables"`)
	// if err != nil {
	// 	log.Println(err)
	// }
	// iptables, err := models.ParseIptables(firewallRule)
	// if err != nil {
	// 	log.Println(err)
	// }
	// fmt.Println(iptables)

	// firewallSetting := `{"host":"vmware-windows", "name":"add firewall test-in"}`
	// models.DeleteFirewallRule(firewallSetting)
	go models.RemoveEntryChannel()
	router := mux.NewRouter().StrictSlash(true)
	credentials := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	methods := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "DELETE", "OPTIONS"})
	origins := handlers.AllowedOrigins([]string{"*"})
	// Login
	router.HandleFunc("/login", routes.Login).Methods("POST", "OPTIONS")

	// SSH Connection
	router.HandleFunc("/sshconnection", routes.SSHCopyKey).Methods("POST", "OPTIONS")
	router.HandleFunc("/sshconnection/{id}/test", routes.TestSSHConnection).Methods("GET", "OPTIONS")
	router.HandleFunc("/sshconnections", routes.GetAllSSHConnection).Methods("GET", "OPTIONS")
	router.HandleFunc("/sshconnection/{id}", routes.SSHConnectionDeleteRoute).Methods("DELETE", "OPTIONS")

	// SSH Key
	router.HandleFunc("/sshkey", routes.AddSSHKey).Methods("POST", "OPTIONS")
	router.HandleFunc("/sshkey/{id}", routes.SSHKeyDeleteRoute).Methods("DELETE", "OPTIONS")
	router.HandleFunc("/sshkeys", routes.GetAllSSHKey).Methods("GET", "OPTIONS")

	// Get PC info
	router.HandleFunc("/systeminfo/{id}", routes.GetSystemInfoRoute).Methods("GET", "OPTIONS")
	router.HandleFunc("/systeminfos", routes.SystemInfoGetAllRoute).Methods("GET", "OPTIONS")
	router.HandleFunc("/receivelog", routes.Receivelog).Methods("POST", "OPTIONS")
	router.HandleFunc("/getufwsettings/{id}", routes.UfwRulesGet).Methods("GET", "OPTIONS")

	//Config client settings
	router.HandleFunc("/addufwrule", routes.AddUfwRule).Methods("POST", "OPTIONS")
	router.HandleFunc("/delufwrule", routes.DeleteUfwRule).Methods("POST", "OPTIONS")

	// Network Function
	router.HandleFunc("/network/defaultip", routes.GetAllDefaultIP).Methods("GET")

	// Package
	router.HandleFunc("/package/install", routes.PackageInstall).Methods("POST")
	router.HandleFunc("/package/remove", routes.PackageRemove).Methods("POST")
	router.HandleFunc("/package/list", routes.PackageListAll).Methods("POST")

	// Host User
	router.HandleFunc("/hostuser/add", routes.HostUserAdd).Methods("POST")
	router.HandleFunc("/hostuser/remove", routes.HostUserRemove).Methods("POST")
	router.HandleFunc("/hostuser/list/{id}", routes.HostUserListAll).Methods("GET")

	// User command history
	// Not finished
	//router.HandleFunc("/history/list/{id}", routes.HistoryListAll).Methods("GET")

	// Event Web
	router.HandleFunc("/eventweb", routes.GetAllEventWeb).Methods("GET")
	router.HandleFunc("/eventweb/delete/all", routes.DeleteAllEventWeb).Methods("GET")

	// Custom API
	router.HandleFunc("/pcs", routes.GetAllPcs).Methods("GET")

	//API management
	router.HandleFunc("/telegrambotoken", routes.AddTelegramBotKey).Methods("POST", "OPTIONS")
	router.HandleFunc("/telegrambotoken", routes.GetTelegramBotKey).Methods("GET", "OPTIONS")

	//Template & Task management
	router.HandleFunc("/templates", routes.AddTemplate).Methods("POST", "OPTIONS")
	router.HandleFunc("/templates", routes.GetAllTemplate).Methods("GET", "OPTIONS")
	router.HandleFunc("/templates/{id}", routes.DeleteTemplate).Methods("DELETE", "OPTIONS")
	router.HandleFunc("/templates/{id}/tasks", routes.GetAllTask).Methods("GET", "OPTIONS")
	router.HandleFunc("/tasks", routes.AddTask).Methods("POST", "OPTIONS")
	router.HandleFunc("/tasks/{id}/logs", routes.GetTaskLog).Methods("GET", "OPTIONS")
	router.HandleFunc("/tasks/{id}/cron/stop", routes.RemoveCronRoute).Methods("GET", "OPTIONS")

	//Log file alert
	router.HandleFunc("/watchfile", routes.WatchFile).Methods("POST", "OPTIONS")

	//Log file serving
	var dir = "/var/log/remotelogs/"
	d := http.Dir(dir)
	fileserver := http.FileServer(d)
	router.PathPrefix("/logs/").Handler(http.StripPrefix("/logs/", fileserver))

	// Web app user
	router.HandleFunc("/wauser/add", routes.AddWebAppUser).Methods("POST")
	router.HandleFunc("/wauser/remove/{id}", routes.DeleteWebAppUser).Methods("GET")
	router.HandleFunc("/wauser/update", routes.UpdateWebAppUser).Methods("POST")
	router.HandleFunc("/wauser/list", routes.ListAllWebAppUser).Methods("GET")
	router.HandleFunc("/wauser/list/{id}", routes.ListWebAppUser).Methods("GET")

	// Network Automation: Vyos
	router.HandleFunc("/vyos/list/{id}", routes.GetInfoVyos).Methods("GET")
	router.HandleFunc("/vyos/config/ip", routes.ConfigIPVyos).Methods("POST")

	//Windows Firewall Settings
	router.HandleFunc("/{id}/firewall/{direction}", routes.GetWindowsFirewall).Methods("GET")

	log.Fatal(http.ListenAndServe(":8081", handlers.CORS(credentials, methods, origins)(router)))
}
