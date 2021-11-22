package routes

import (
	"errors"
	"net/http"
	"time"

	"github.com/wintltr/login-api/auth"
	"github.com/wintltr/login-api/utils"
)

func GetCurrentSystemTime(w http.ResponseWriter, r *http.Request) {

	//Authorization
	isAuthorized, err := auth.CheckAuth(r, []string{"admin", "user"})
	if err != nil {
		utils.ERROR(w, http.StatusUnauthorized, errors.New("invalid token").Error())
		return
	}
	if !isAuthorized {
		utils.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized").Error())
		return
	}

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
