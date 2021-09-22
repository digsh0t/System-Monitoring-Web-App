package routes

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/bitly/go-simplejson"
	"github.com/wintltr/login-api/auth"
	"github.com/wintltr/login-api/models"
	"github.com/wintltr/login-api/utils"
)

func InventoryGroupAdd(w http.ResponseWriter, r *http.Request) {

	// Authorization
	isAuthorized, err := auth.CheckAuth(r, []string{"admin"})
	if err != nil {
		utils.ERROR(w, http.StatusUnauthorized, errors.New("invalid token").Error())
		return
	}
	if !isAuthorized {
		utils.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized").Error())
		return
	}

	// Retrieve Json Format
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.ERROR(w, http.StatusUnauthorized, errors.New("fail to parse json").Error())
		return
	}

	var inventoryGroup models.InventoryGroup
	json.Unmarshal(reqBody, &inventoryGroup)

	status, err := models.InventoryGroupAdd(inventoryGroup)

	// Return Json
	returnJson := simplejson.New()
	returnJson.Set("Status", status)
	if err != nil {
		returnJson.Set("Error", err.Error())
	} else {
		returnJson.Set("Error", err)
	}
	utils.JSON(w, http.StatusBadRequest, returnJson)

	// Write Event Web
	description := "Inventory \"" + inventoryGroup.GroupName + "\" added successfully"
	_, err = models.WriteWebEvent(r, "InventoryGroup", description)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to write event").Error())
		return
	}
}
