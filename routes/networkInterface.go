package routes

import (
	"net/http"

	"github.com/bitly/go-simplejson"
	"github.com/wintltr/login-api/event"
	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/utils"
)

func GetAllDefaultIP(w http.ResponseWriter, r *http.Request) {

	defaultIPList, err := models.GetAllDefaultIP()
	if err != nil {
		utils.JSON(w, http.StatusBadRequest, defaultIPList)
		return
	}

	// Write Event Web
	returnJson := simplejson.New()
	description := "Show all network interface of all clients"
	_, err = event.WriteWebEvent(r, "Network", description)

	if err != nil {
		returnJson.Set("Status", false)
		returnJson.Set("Error", "Fail to write web event")
		utils.JSON(w, http.StatusBadRequest, returnJson)
		return
	}
	utils.JSON(w, http.StatusOK, defaultIPList)

}
