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
	// sshKey, err := models.GetSSHKeyFromId(14)
	// if err != nil {
	// 	log.Println(err)
	// }
	// err = sshKey.WriteKeyToFile("tmp/private_key")
	//err := models.RemoveFile("tmp/private_key")
	// if err != nil {
	// 	log.Println(err)
	// }

	// firewallSetting := `{"host":"vmware-windows", "name":"add firewall test-in"}`
	//models.DeleteFirewallRule(firewallSetting)
	// sshConnection, err := models.GetSSHConnectionFromId(51)
	// if err != nil {
	// 	log.Println(err)
	// }
	// sid := "S-1-5-21-1572063403-3487170947-126735497-1000"
	// var keyList []models.RegistryKey
	// keyList = append(keyList, models.RegistryKey{Data: "1", Path: "Disables all Control Panel programs and the PC settings app"})
	// var userList []models.ClientUser
	// userList = append(userList, models.ClientUser{Username: "Administrator"}, models.ClientUser{Username: "guest"}, models.ClientUser{Username: "TRILX"})
	// output, err := sshConnection.CheckSSHConnectionExist()
	// if err != nil {
	// 	log.Println(err)
	// }
	// log.Println(output)

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

	// SSH Key
	router.HandleFunc("/sshkey", routes.AddSSHKey).Methods("POST", "OPTIONS")
	router.HandleFunc("/sshkey/{id}", routes.SSHKeyDeleteRoute).Methods("DELETE", "OPTIONS")
	router.HandleFunc("/sshkeys", routes.GetAllSSHKey).Methods("GET", "OPTIONS")

	// Inventory Group
	router.HandleFunc("/inventory/group/add", routes.InventoryGroupAdd).Methods("POST")
	router.HandleFunc("/inventory/group/list", routes.InventoryGroupList).Methods("GET")
	router.HandleFunc("/inventory/group/delete/{id}", routes.InventoryGroupDelete).Methods("DELETE")
	router.HandleFunc("/sshconnections/list/nogroup", routes.GetAllSSHConnectionNoGroup).Methods("GET", "OPTIONS")
	router.HandleFunc("/inventory/group/addclient", routes.InventoryGroupAddClient).Methods("POST")
	router.HandleFunc("/inventory/group/deleteclient", routes.InventoryGroupDeleteClient).Methods("POST")
	router.HandleFunc("/inventory/group/listclient/{groupid}", routes.InventoryGroupListClient).Methods("GET")

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
	router.HandleFunc("/linux/user/remove", routes.LinuxClientUserRemove).Methods("DELETE")
	router.HandleFunc("/linux/user/list/{id}", routes.LinuxClientUserListAll).Methods("GET")

	// Linux Client Group
	router.HandleFunc("/linux/group/add", routes.LinuxClientGroupAdd).Methods("POST")
	router.HandleFunc("/linux/group/remove", routes.LinuxClientGroupRemove).Methods("DELETE")
	router.HandleFunc("/linux/group/list", routes.LinuxClientGroupListAll).Methods("POST")

	// Linux Client Iptables
	router.HandleFunc("/linux/iptables/list/{id}", routes.LinuxClientIptablesListAll).Methods("GET")
	router.HandleFunc("/linux/iptables/add", routes.LinuxClientIptablesAdd).Methods("POST")
	router.HandleFunc("/linux/iptables/remove", routes.LinuxClientIptablesRemove).Methods("DELETE")

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
	router.HandleFunc("/templates/argument/{id}", routes.GetTemplateArgument).Methods("GET", "OPTIONS")
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
	//router.HandleFunc("/vyos/listconfig/{id}", routes.GetInfoConfigVyos).Methods("GET")
	router.HandleFunc("/vyos/list/{id}", routes.GetInfoVyos).Methods("GET")
	router.HandleFunc("/vyos/config/ip", routes.ConfigIPVyos).Methods("POST")
	router.HandleFunc("/vyos/list", routes.ListAllVyOS).Methods("GET")
	router.HandleFunc("/vyos/logs", routes.ListLogsVyos).Methods("GET")

	// Network Automation: Cisco
	router.HandleFunc("/cisco/list", routes.ListAllCisco).Methods("GET")
	router.HandleFunc("/cisco/listconfig/{id}", routes.GetInfoConfigCisco).Methods("GET")
	router.HandleFunc("/cisco/listinterface/{id}", routes.GetInfoInterfaceCisco).Methods("GET")
	router.HandleFunc("/cisco/config/ip", routes.ConfigIPCisco).Methods("POST")
	router.HandleFunc("/cisco/config/staticroute", routes.ConfigStaticRouteCisco).Methods("POST")
	router.HandleFunc("/cisco/testping", routes.TestPingCisco).Methods("POST")
	router.HandleFunc("/cisco/logs", routes.ListLogsCisco).Methods("GET")
	router.HandleFunc("/cisco/traffic", routes.GetTrafficCisco).Methods("GET")

	//Windows Firewall Settings
	router.HandleFunc("/{id}/firewall/{direction}", routes.GetWindowsFirewall).Methods("OPTIONS", "GET")
	router.HandleFunc("/firewall", routes.AddWindowsFirewall).Methods("OPTIONS", "POST")
	router.HandleFunc("/firewall", routes.RemoveWindowsFirewallRule).Methods("OPTIONS", "DELETE")
	router.HandleFunc("/{id}/openconnection", routes.GetWindowsOpenConnection).Methods("OPTIONS", "GET")

	//Windows Programs Management
	router.HandleFunc("/{id}/programs", routes.GetWindowsInstalledProgram).Methods("GET")
	router.HandleFunc("/programs", routes.InstallWindowsProgram).Methods("POST")
	router.HandleFunc("/programs", routes.RemoveWindowsProgram).Methods("DELETE")

	//Add new ssh connection
	router.HandleFunc("/newsshconnection", routes.AddNewSSHConnection).Methods("POST")

	//Windows Event Log
	router.HandleFunc("/windows/eventlog", routes.GetWindowsEventLogs).Methods("GET")
	router.HandleFunc("/windows/eventlog/detail", routes.GetDetailWindowsEventLog).Methods("GET")

	//Windows Local Users Management
	router.HandleFunc("/{id}/localuser", routes.GetWindowsLocalUser).Methods("OPTIONS", "GET")
	router.HandleFunc("/{id}/localuser/changepassword", routes.ChangeWindowsLocalUserPassword).Methods("OPTIONS", "POST")
	//router.HandleFunc("/{id}/localuser/{username}/enabled", routes.GetWindowsUserEnableStatus).Methods("OPTIONS", "GET")
	router.HandleFunc("/{id}/localuser/{username}/enabled/{is_enabled}", routes.ChangeWindowsEnabledStatus).Methods("OPTIONS", "PUT")
	router.HandleFunc("/{id}/localuser/{username}/groups", routes.GetWindowsGroupListOfUser).Methods("OPTIONS", "GET")
	router.HandleFunc("/localuser/groups", routes.ReplaceWindowsGroupOfUser).Methods("OPTIONS", "POST")
	router.HandleFunc("/localuser", routes.AddNewWindowsLocalUser).Methods("OPTIONS", "POST")
	router.HandleFunc("/localuser", routes.DeleteWindowsUser).Methods("OPTIONS", "DELETE")
	router.HandleFunc("/{id}/loggedinusers", routes.GetLoggedInUsers).Methods("OPTIONS", "GET")
	router.HandleFunc("/{id}/loggedinusers/{session_id}", routes.KillWindowsLogonSession).Methods("OPTIONS", "DELETE")
	router.HandleFunc("/{id}/loggedinusers/{username}/appexecutionhistory", routes.GetWindowsLogonAppExecutionHistory).Methods("OPTIONS", "GET")

	//Windows Local Group Management
	router.HandleFunc("/{id}/localusergroup", routes.GetWindowsLocalUserGroup).Methods("OPTIONS", "GET")
	router.HandleFunc("/localusergroup", routes.AddNewWindowsGroup).Methods("OPTIONS", "POST")
	router.HandleFunc("/localusergroup", routes.RemoveWindowsGroup).Methods("OPTIONS", "DELETE")

	//Install guide
	router.HandleFunc("/manual", routes.GetInstallManual).Methods("OPTIONS", "GET")

	//Get Windows Processes
	router.HandleFunc("/{id}/processes", routes.GetWindowsProcesses).Methods("OPTIONS", "GET")
	router.HandleFunc("/{id}/processes/{pid}", routes.KillWindowsProcess).Methods("OPTIONS", "DELETE")

	//Get Windows Sys Info
	router.HandleFunc("/{id}/osversion", routes.GetOSVersion).Methods("OPTIONS", "GET")
	router.HandleFunc("/{id}/cpuinfo", routes.GetCPUInfo).Methods("OPTIONS", "GET")
	router.HandleFunc("/{id}/interfaces", routes.GetInterfaceList).Methods("OPTIONS", "GET")
	router.HandleFunc("/{id}/connectivity", routes.GetConnectivityInfo).Methods("OPTIONS", "GET")

	//Windows Service
	router.HandleFunc("/{id}/services", routes.GetWindowsServiceList).Methods("OPTIONS", "GET")
	router.HandleFunc("/{id}/services/{service_name}/{service_state}", routes.ChangeWindowsServiceState).Methods("OPTIONS", "PUT")

	//Windows Policy
	router.HandleFunc("/{id}/policies/{sid}/explorer", routes.GetWindowsExplorerPolicy).Methods("OPTIONS", "GET")
	router.HandleFunc("/{id}/policies/{sid}/explorer", routes.ChangeWindowsExplorerPolicy).Methods("OPTIONS", "POST")
	router.HandleFunc("/{id}/policies/{sid}/prohibitedprograms", routes.GetWindowsUserProhibitedProgramsPolicy).Methods("OPTIONS", "GET")
	router.HandleFunc("/{id}/policies/{sid}/prohibitedprograms", routes.ChangeWindowsUserProhibitedProgramPolicy).Methods("OPTIONS", "POST")
	router.HandleFunc("/{id}/passwordpolicies", routes.GetWindowsPasswordPolicy).Methods("OPTIONS", "GET")
	router.HandleFunc("/{id}/passwordpolicies", routes.ChangeWindowsPasswordPolicy).Methods("OPTIONS", "PUT")

	//2FA QR managements
	router.HandleFunc("/qr/on", routes.GenerateQR).Methods("OPTIONS", "GET")
	router.HandleFunc("/qr/off", routes.TurnOff2FARoute).Methods("OPTIONS", "POST")
	router.HandleFunc("/qr/verify", routes.VerifyQR).Methods("OPTIONS", "POST")
	router.HandleFunc("/qr/on/verify", routes.VerifyQRSettingsRoute).Methods("OPTIONS", "POST")

	//Syslog
	router.HandleFunc("/{id}/syslog/{date}", routes.GetSysLogFilesRoute).Methods("OPTIONS", "GET")

	log.Fatal(http.ListenAndServe(":8081", handlers.CORS(credentials, methods, origins)(router)))
}
