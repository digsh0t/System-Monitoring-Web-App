package routes

import (
	"errors"
	"net/http"

	"github.com/wintltr/login-api/event"
	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/utils"
)

func GetAllDefaultIP(w http.ResponseWriter, r *http.Request) {

	defaultIPList, err := models.GetAllDefaultIP()
	var eventStatus string
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("Fail to retrieve ip ").Error())
		eventStatus = "failed"
	} else {
		utils.JSON(w, http.StatusOK, defaultIPList)
		eventStatus = "successully"
	}

	// Write Event Web
	description := "List all ip of clients " + eventStatus
	_, err = event.WriteWebEvent(r, "Login", description)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("Fail to write event").Error())
		return
	}

}
