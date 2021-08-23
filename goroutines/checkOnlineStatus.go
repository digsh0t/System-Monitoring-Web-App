package goroutines

import (
	"fmt"
	"time"

	"github.com/wintltr/login-api/models"
)

func CheckClientOnlineStatusGour() {
	onlineStatusesChan := make(chan []models.OnlineStatus)
	sshConnection, _ := models.GetAllSSHConnection()
	for {
		time.Sleep(100 * time.Second)
		go func() { onlineStatusesChan <- models.CheckOnlineStatus(sshConnection) }()
		onlineStatuses := <-onlineStatusesChan
		fmt.Println(onlineStatuses)
	}
}
