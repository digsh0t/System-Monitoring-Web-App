package routes

import (
	"net/http"
	"time"

	"github.com/wintltr/login-api/utils"
)

func GetCurrentSystemTime(w http.ResponseWriter, r *http.Request) {

	type timezone struct {
		CurrentTime string `json:"current_time"`
		Zone        string `json:"zone"`
		Offset      int    `json:"offset"`
	}
	var tz timezone
	tz.CurrentTime = time.Now().Local().String()
	tz.Zone, tz.Offset = time.Now().Zone()
	utils.JSON(w, http.StatusOK, tz)
}
