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
	// sshConnection, err := models.GetSSHConnectionFromId(33)
	// if err != nil {
	// 	log.Println(err)
	// }
	// userList, err := sshConnection.GetLocalUsers()
	// if err != nil {
	// 	log.Println(err)
	// }
	// log.Println(userList)

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
	router.HandleFunc("/sshconnections/{ostype}", routes.GetAllSSHConnection).Methods("GET", "OPTIONS")
	router.HandleFunc("/sshconnections", routes.GetAllSSHConnection).Methods("GET", "OPTIONS")
	router.HandleFunc("/sshconnection/{id}", routes.SSHConnectionDeleteRoute).Methods("DELETE", "OPTIONS")
	router.HandleFunc("/sshconnections", routes.GetAllSSHConnectionNoGroup).Methods("GET", "OPTIONS")

	// SSH Key
	router.HandleFunc("/sshkey", routes.AddSSHKey).Methods("POST", "OPTIONS")
	router.HandleFunc("/sshkey/{id}", routes.SSHKeyDeleteRoute).Methods("DELETE", "OPTIONS")
	router.HandleFunc("/sshkeys", routes.GetAllSSHKey).Methods("GET", "OPTIONS")

	// Inventory Group
	router.HandleFunc("/inventory/group/add", routes.InventoryGroupAdd).Methods("POST")

	// Get PC info
	router.HandleFunc("/systeminfo/{id}", routes.GetSystemInfoRoute).Methods("GET", "OPTIONS")
	router.HandleFunc("/systeminfos", routes.SystemInfoGetAllRoute).Methods("GET", "OPTIONS")
	router.HandleFunc("/receivelog", routes.Receivelog).Methods("POST", "OPTIONS")

	// Network Function
	router.HandleFunc("/network/defaultip", routes.GetAllDefaultIP).Methods("GET")

	// Package
	router.HandleFunc("/linux/package/install", routes.PackageInstall).Methods("POST")
	router.HandleFunc("/linux/package/remove", routes.PackageRemove).Methods("POST")
	router.HandleFunc("/linux/package/list", routes.PackageListAll).Methods("POST")

	// Linux Client User
	router.HandleFunc("/linux/user/add", routes.LinuxClientUserAdd).Methods("POST")
	router.HandleFunc("/linux/user/remove", routes.LinuxClientUserRemove).Methods("POST")
	router.HandleFunc("/linux/user/list/{id}", routes.LinuxClientUserListAll).Methods("GET")

	// Linux Client Group
	router.HandleFunc("/linux/group/add", routes.LinuxClientGroupAdd).Methods("POST")
	router.HandleFunc("/linux/group/remove", routes.LinuxClientGroupRemove).Methods("POST")
	router.HandleFunc("/linux/group/list/{id}", routes.LinuxClientGroupListAll).Methods("GET")

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
	router.HandleFunc("/vyos/list", routes.ListAllVyOS).Methods("GET")

	//Windows Firewall Settings
	router.HandleFunc("/{id}/firewall/{direction}", routes.GetWindowsFirewall).Methods("OPTIONS", "GET")
	router.HandleFunc("/firewall", routes.AddWindowsFirewall).Methods("OPTIONS", "POST")
	router.HandleFunc("/firewall", routes.RemoveWindowsFirewallRule).Methods("OPTIONS", "DELETE")

	//Windows Programs Management
	router.HandleFunc("/{id}/programs", routes.GetWindowsInstalledProgram).Methods("GET")
	router.HandleFunc("/programs", routes.InstallWindowsProgram).Methods("POST")
	router.HandleFunc("/programs", routes.RemoveWindowsProgram).Methods("DELETE")

	//Add new ssh connection
	router.HandleFunc("/newsshconnection", routes.AddNewSSHConnection).Methods("POST")

	//Windows Local Users Management
	router.HandleFunc("/{id}/localuser", routes.GetWindowsLocalUser).Methods("OPTIONS", "GET")
	router.HandleFunc("/localuser", routes.AddNewWindowsLocalUser).Methods("OPTIONS", "POST")
	router.HandleFunc("/localuser", routes.DeleteWindowsUser).Methods("OPTIONS", "DELETE")

	//Windows Local Group Management
	router.HandleFunc("/{id}/localusergroup", routes.GetWindowsLocalUserGroup).Methods("OPTIONS", "GET")
	router.HandleFunc("/localusergroup", routes.AddNewWindowsGroup).Methods("OPTIONS", "POST")
	router.HandleFunc("/localusergroup", routes.RemoveWindowsGroup).Methods("OPTIONS", "DELETE")

	log.Fatal(http.ListenAndServe(":8081", handlers.CORS(credentials, methods, origins)(router)))
}
