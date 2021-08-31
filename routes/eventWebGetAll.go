package routes

import (
	"errors"
	"net/http"

	"github.com/bitly/go-simplejson"
	"github.com/wintltr/login-api/auth"
	"github.com/wintltr/login-api/event"
	"github.com/wintltr/login-api/utils"
)

func GetAllEventWeb(w http.ResponseWriter, r *http.Request) {

	//Authorization
	isAuthorized, err := auth.CheckAuth(r, []string{"admin"})
	if err != nil {
		utils.ERROR(w, http.StatusUnauthorized, errors.New("invalid token").Error())
		return
	}
	if !isAuthorized {
		utils.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized").Error())
		return
	}

	eventWebList, err := event.GetAllEventWeb()

	// Write Event Web
	returnJson := simplejson.New()
	description := "Display all event web"
	_, err = event.WriteWebEvent(r, "event", description)
	if err != nil {
		returnJson.Set("Status", false)
		returnJson.Set("Error", "Fail to write web event")
		utils.JSON(w, http.StatusBadRequest, returnJson)
		return
	}

	// Return json
	if err != nil {

		utils.ERROR(w, http.StatusBadRequest, errors.New("Fail to get all pc").Error())
		return
	}

	utils.JSON(w, http.StatusOK, eventWebList)
}
