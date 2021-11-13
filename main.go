package main

import (
	"github.com/wintltr/login-api/cmd"
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

	// sshConnection, err := models.GetSSHConnectionFromId(59)
	// if err != nil {
	// 	log.Println(err)
	// }
	// key, err := sshConnection.GetLinuxUsersLastLogin()
	// if err != nil {
	// 	log.Println(err)
	// }
	// log.Println(key)
	// for _, index := range key {
	// 	fmt.Print(index.Username + " ")
	// 	fmt.Print(index.IsEnabled)
	// 	fmt.Print(" " + index.LastLogon)
	// 	fmt.Println(index.Type)
	// }
	// key, err = sshConnection.GetWindowsVmwareProductKey()
	// if err != nil {
	// 	log.Println(err)
	// }
	// fmt.Println(key)
	//err := models.ExportReport("hello.pdf")
	//if err != nil {
	//	log.Println(err.Error())
	//}
	// sI := models.SmtpInfo{EmailSender: "noti.lthmonitor@gmail.com", EmailPassword: "Lethihang123", SMTPHost: "smtp.gmail.com", SMTPPort: "587"}
	// sI.SendReportMail("./10-11-2021-report.pdf", []string{"trilxse140935@fpt.edu.vn"})

	cmd.Execute()
}
