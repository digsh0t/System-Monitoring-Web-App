package routes

import (
	"net/http"

	"github.com/bitly/go-simplejson"
	"github.com/wintltr/login-api/auth"
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
	id, err := auth.ExtractUserId(r)
	if err != nil {
		returnJson.Set("Status", false)
		returnJson.Set("Error", "Fail to get id of creator")
		utils.JSON(w, http.StatusBadRequest, returnJson)
		return
	}

	var eventWeb event.EventWeb = event.EventWeb{
		EventWebType:        "Network",
		EventWebDescription: "Show all network interface of all clients ",
		EventWebCreatorId:   id,
	}
	_, err = eventWeb.WriteWebEvent()

	if err != nil {
		returnJson.Set("Status", false)
		returnJson.Set("Error", "Fail to write web event")
		utils.JSON(w, http.StatusBadRequest, returnJson)
		return
	}
	utils.JSON(w, http.StatusOK, defaultIPList)

}
