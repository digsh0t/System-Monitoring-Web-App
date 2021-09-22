package routes

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/bitly/go-simplejson"
	"github.com/gorilla/mux"
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
	var statusCode int
	returnJson.Set("Status", status)
	if err != nil {
		returnJson.Set("Error", err.Error())
		statusCode = http.StatusBadRequest
	} else {
		returnJson.Set("Error", err)
		statusCode = http.StatusOK
	}

	utils.JSON(w, statusCode, returnJson)

	// Write Event Web
	description := "Inventory \"" + inventoryGroup.GroupName + "\" added successfully"
	_, err = models.WriteWebEvent(r, "InventoryGroup", description)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to write event").Error())
		return
	}
}

// Get Inventory_group from DB
func InventoryGroupList(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	if r.Method == "OPTIONS" {
		//CORS
		// return "OKOK"
		json.NewEncoder(w).Encode("OKOK")
		return
	}

	isAuthorized, err := auth.CheckAuth(r, []string{"admin", "user"})
	if err != nil {
		utils.ERROR(w, http.StatusUnauthorized, errors.New("invalid token").Error())
		return
	}
	if !isAuthorized {
		utils.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized").Error())
		return
	}

	inventGroupList, err := models.GetAllInventoryGroup()
	if err != nil {
		utils.JSON(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, inventGroupList)

}

// Delete Inventory_group from DB
func InventoryGroupDelete(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	if r.Method == "OPTIONS" {
		//CORS
		// return "OKOK"
		json.NewEncoder(w).Encode("OKOK")
		return
	}

	isAuthorized, err := auth.CheckAuth(r, []string{"admin", "user"})
	if err != nil {
		utils.ERROR(w, http.StatusUnauthorized, errors.New("invalid token").Error())
		return
	}
	if !isAuthorized {
		utils.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized").Error())
		return
	}

	vars := mux.Vars(r)
	groupId, _ := strconv.Atoi(vars["id"])

	result, err := models.InventoryGroupDelete(groupId)

	// Return Json
	returnJson := simplejson.New()
	var statusCode int
	returnJson.Set("Status", result)
	if err != nil {
		returnJson.Set("Error", err.Error())
		statusCode = http.StatusBadRequest
	} else {
		returnJson.Set("Error", err)
		statusCode = http.StatusOK
	}

	utils.JSON(w, statusCode, returnJson)

}

// Add client to inventory Group
func InventoryGroupAddClient(w http.ResponseWriter, r *http.Request) {

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

	status, err := models.InventoryGroupAddClient(inventoryGroup)

	// Return Json
	var statusCode int
	returnJson := simplejson.New()
	returnJson.Set("Status", status)
	if err != nil {
		returnJson.Set("Error", err.Error())
		statusCode = http.StatusBadRequest
	} else {
		returnJson.Set("Error", err)
		statusCode = http.StatusOK
	}
	utils.JSON(w, statusCode, returnJson)

	// Write Event Web
	description := "Inventory \"" + inventoryGroup.GroupName + "\" added successfully"
	_, err = models.WriteWebEvent(r, "InventoryGroup", description)
	if err != nil {
		utils.ERROR(w, http.StatusBadRequest, errors.New("fail to write event").Error())
		return
	}
}
